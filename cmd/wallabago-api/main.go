package main

import (
	"log/slog"
	"os"

	"github.com/andriihomiak/wallabago/internal/http"
)

func main() {
	server, err := http.NewServer()
	if err != nil {
		slog.Error("failed to create server", "cause", err)
	}
	slog.Error("Stopped server", "cause", server.Start())
	os.Exit(1)
}
