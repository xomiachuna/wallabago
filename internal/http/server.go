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

	"github.com/andriihomiak/wallabago/internal/app"
	"github.com/pkg/errors"
)

type Server struct {
	app app.Wallabago
}

func NewServer(cfg app.Config) (*Server, error) {
	wallabago, err := app.NewWallabago(context.Background(), &cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &Server{
		app: *wallabago,
	}, nil
}

func (s *Server) Start() error {
	// listen for interrupt signal
	rootCtx, stopListeningForInterrupt := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGSTOP,
	)
	defer stopListeningForInterrupt()

	rootHandler := http.NewServeMux()

	s.app.RegisterHandlers(rootHandler)

	server := &http.Server{
		Addr:    s.app.Addr(),
		Handler: rootHandler,
		// the context here is the one cancellable by interrupt
		BaseContext: func(_ net.Listener) context.Context { return rootCtx },
		// preventing slowloris attack as advised by gosec - increase if needed
		// can also be handled in an upstream server
		ReadHeaderTimeout: time.Millisecond * 1000,
	}

	// perform bootstrap
	err := s.app.Prepare(rootCtx)
	if err != nil {
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
		// shutdown due to server start failure

		slog.Warn("Server start failed", "cause", err)
		shutdownCtx := context.TODO()
		appShutdownErr := s.app.Shutdown(shutdownCtx)
		return stderrors.Join(err, appShutdownErr)

	case <-rootCtx.Done():
		// shutdown due to interrupt

		// todo: add logic for graceful handling of readiness probes
		slog.Warn("Received interrupt, shutting down server", "cause", rootCtx.Err())
		// this allows for forceful termination using second interrupt
		stopListeningForInterrupt()
		// cant reuse the root context here as it is honored by shutdown and
		// communicates the timeout for graceful shutdown
		shutdownCtx := context.TODO()
		serverShutdownErr := server.Shutdown(shutdownCtx)
		appShutdownErr := s.app.Shutdown(shutdownCtx)
		return stderrors.Join(err, appShutdownErr, serverShutdownErr)
	}
}
