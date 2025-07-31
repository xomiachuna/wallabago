package http

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	stderrors "errors"

	"github.com/andriihomiak/wallabago/internal/storage"

	bootstrapManagers "github.com/andriihomiak/wallabago/internal/bootstrap/managers"
	"github.com/andriihomiak/wallabago/internal/database"
	"github.com/andriihomiak/wallabago/internal/http/handlers"
	"github.com/andriihomiak/wallabago/internal/http/middleware"
	identityHandlers "github.com/andriihomiak/wallabago/internal/identity/handlers"
	identityStorage "github.com/andriihomiak/wallabago/internal/identity/storage"
	"github.com/andriihomiak/wallabago/internal/instrumentation"
	"github.com/pkg/errors"
)

type App struct {
	instrumentationEnabled bool
	addr                   string
	identityPool           *sql.DB
	appPool                *sql.DB
}

func getDefaultMiddleware() middleware.Middleware {
	return middleware.NewChain(
		middleware.NewOtelHTTPMiddleware(),
	)
}

func (a *App) newRootHandler() http.Handler {
	mux := http.NewServeMux()

	oauth2 := identityHandlers.NewOAuth2HandlerFromDBPool(a.identityPool)
	mux.HandleFunc("POST /oauth/v2/token", oauth2.TokenEndpoint)

	service := handlers.Index{}
	mux.HandleFunc("/", service.Index)

	globalMiddleware := getDefaultMiddleware()
	return globalMiddleware.Wrap(mux)
}

func (a *App) Start() error {
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

	if a.instrumentationEnabled {
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

	// todo: pass db url as a parameter
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

	// todo: use separate db pools?
	a.identityPool = dbPool
	a.appPool = dbPool

	rootHandler := a.newRootHandler()
	server := &http.Server{
		Addr:    a.addr,
		Handler: rootHandler,
		// the context here is the one cancellable by interrupt
		BaseContext: func(_ net.Listener) context.Context { return rootCtx },
		// preventing slowloris attack as advised by gosec - increase if needed
		// can also be handled in an upstream server
		ReadHeaderTimeout: time.Millisecond * 1000,
	}

	// perform bootstrap
	bootstrapManager := bootstrapManagers.NewManager(
		identityStorage.NewCodegenStorage(a.identityPool),
		storage.NewGeneratedBootstrapSQLStorage(a.appPool),
	)

	err = bootstrapManager.Bootstrap(rootCtx)
	if err != nil {
		slog.ErrorContext(rootCtx, "Failed to bootstrap", "cause", err.Error())
		return errors.WithStack(err)
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
	_, enableOtel := os.LookupEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	addr := ":8080"
	port, ok := os.LookupEnv("WALLABAGO_PORT")
	if ok {
		// todo: verify port is int
		addr = fmt.Sprintf(":%s", port)
	}
	return &App{
		instrumentationEnabled: enableOtel,
		addr:                   addr,
	}, nil
}
