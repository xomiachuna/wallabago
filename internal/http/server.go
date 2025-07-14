package http

import "log/slog"

type Server struct{}

func (s *Server) Start() error {
	slog.Info("Starting server")
	return nil
}

func NewServer() (*Server, error) {
	return &Server{}, nil
}
