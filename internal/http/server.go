package http

import (
	"log/slog"

	"github.com/andriihomiak/wallabago/internal/instrumentation"
)

type Config struct {
	Port            int
	Instrumentation instrumentation.Config
}

type Server struct {
	Config Config
}

func (s *Server) Start() error {
	instrumentation.Instrument(s.Config.Instrumentation)
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
