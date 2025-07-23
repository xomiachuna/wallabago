package http

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	stderrors "errors"

	"github.com/andriihomiak/wallabago/internal/database"
	"github.com/andriihomiak/wallabago/internal/http/handlers"
	"github.com/andriihomiak/wallabago/internal/http/middleware"
	"github.com/andriihomiak/wallabago/internal/instrumentation"
	"github.com/pkg/errors"
)

type App struct {
	instrumentationEnabled bool
}

func getDefaultMiddleware() middleware.Middleware {
	return middleware.NewChain(
		middleware.NewOtelHTTPMiddleware(),
	)
}

func newMux(querier database.Querier) *http.ServeMux {
	innerMux := http.NewServeMux()
	service := handlers.NewService(querier)
	innerMux.HandleFunc("/", service.Index)
	globalMiddleware := getDefaultMiddleware()
	outerMux := http.NewServeMux()
	outerMux.Handle("/", globalMiddleware.Wrap(innerMux))
	return outerMux
}

func (wb *App) Start() error {
	// listen for interrupt signal
	rootCtx, stopListeningForInterrupt := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGSTOP,
	)
	defer stopListeningForInterrupt()

	var shutdownOtel func(context.Context) error
	var err error

	if wb.instrumentationEnabled {
		shutdownOtel, err = instrumentation.SetupOtelSDK(rootCtx)
		if err != nil {
			return errors.Wrap(err, "Failed to setup otel")
		}
	} else {
		slog.Warn("Otel instrumentation is not enabled")
		shutdownOtel = func(_ context.Context) error {
			slog.Warn("Otel instrumentation is not enabled, nothing to cleanup")
			return nil
		}
	}

	dbPool, err := database.NewDBPool(rootCtx, os.Getenv("DB"))
	if err != nil {
		return errors.Wrap(err, "Failed to connect to db")
	}

	shutdownDb := func() error {
		slog.Warn("Closing db pool")
		closeErr := dbPool.Close()
		slog.Warn("FB pool closed")
		return closeErr
	}

	querier := database.New(dbPool)

	mux := newMux(querier)
	server := &http.Server{
		// TODO: pass from outside?
		Addr:    ":8080",
		Handler: mux,
		// the context here is the one cancellable by interrupt
		BaseContext: func(_ net.Listener) context.Context { return rootCtx },
		// preventing slowloris attack as advised by gosec - increase if needed
		// can also be handled in an upstream server
		ReadHeaderTimeout: time.Millisecond * 1000,
	}

	// start the server and wait for server error
	serveErr := make(chan error, 1)
	go func(serveErr chan<- error) {
		slog.Info("Starting server", "addr", server.Addr)
		serveErr <- server.ListenAndServe()
	}(serveErr)

	// we stop either due to server error or an interrupt
	select {
	case err := <-serveErr:
		// shutdown due to server failure
		slog.Warn("Server failed", "cause", err)
		slog.Warn("Shutting down otel")
		otelShutdownCtx := context.TODO()
		otelShutdownErr := errors.Wrap(shutdownOtel(otelShutdownCtx), "errors during otel shutdown")
		// shutdown db as well
		return stderrors.Join(shutdownDb(), otelShutdownErr, err)

	case <-rootCtx.Done():
		// todo: add logic for readiness probes
		// shutdown due to interrupt
		slog.Warn("Received interrupt, shutting down server", "cause", rootCtx.Err())
		// this allows for forceful termination using second interrupt
		stopListeningForInterrupt()
		// cant reuse the context here as it is honored by shutdown and
		// communicates the timeout for graceful shutdown
		serverShutdownCtx := context.TODO()
		serverShutdownErr := errors.Wrap(server.Shutdown(serverShutdownCtx), "error during server shutdown")
		slog.Warn("Server shut down finished", "errorDuringShutdown", serverShutdownErr)

		slog.Warn("Shutting down otel")
		otelShutdownCtx := context.TODO()
		otelShutdownErr := errors.Wrap(shutdownOtel(otelShutdownCtx), "error during otel shutdown")
		slog.Warn("Otel shut down finished", "errorDuringShutdown", otelShutdownErr)

		// shutdown db as well
		return stderrors.Join(shutdownDb(), otelShutdownErr, serverShutdownErr)
	}
}

func NewApp() (*App, error) {
	return &App{
		instrumentationEnabled: true,
	}, nil
}
