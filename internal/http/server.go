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

	"github.com/andriihomiak/wallabago/internal/http/handlers"
	"github.com/andriihomiak/wallabago/internal/instrumentation"
	"github.com/pkg/errors"
)

type App struct {
	mux *http.ServeMux
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Index)
	mux.HandleFunc("/oauth/v2/token", handlers.TokenEndpoint)
	return mux
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

	shutdownOtel, err := instrumentation.SetupOtelSDK(rootCtx)
	if err != nil {
		return errors.Wrap(err, "Failed to setup otel")
	}

	server := &http.Server{
		// TODO: pass from outside?
		Addr:    ":8080",
		Handler: newMux(),
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
		return stderrors.Join(otelShutdownErr, err)

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

		return stderrors.Join(otelShutdownErr, serverShutdownErr)
	}
}

func NewApp() (*App, error) {
	return &App{
		mux: newMux(),
	}, nil
}
