package http

import (
	"context"
	"log/slog"

	// enable instrumentation.
	"github.com/andriihomiak/wallabago/internal/instrumentation"
	"github.com/pkg/errors"
)

type Config struct {
	Port int
}

type Server struct {
	Config Config
}

func (s *Server) Start() error {
	ctx := context.Background()
	shutdownOtel, err := instrumentation.SetupOtelSDK(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to setup otel")
	}

	slog.Info("Starting server")
	slog.Warn("Stopping server")
	otelShutdownErr := shutdownOtel(ctx)
	if otelShutdownErr != nil {
		return errors.Wrap(otelShutdownErr, "errors during otel shutdown")
	}
	return nil
}

func NewServer() (*Server, error) {
	return &Server{
		Config{
			Port: 7080,
		},
	}, nil
}
