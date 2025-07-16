package http

import (
	"log/slog"
	// enable instrumentation.
	_ "github.com/andriihomiak/wallabago/internal/instrumentation"
)

type Config struct {
	Port int
}

type Server struct {
	Config Config
}

func (s *Server) Start() error {
	slog.Info("Starting server")
	return nil
}

func NewServer() (*Server, error) {
	return &Server{
		Config{
			Port: 7080,
		},
	}, nil
}
