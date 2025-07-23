package main

import (
	"log/slog"
	"os"

	"github.com/andriihomiak/wallabago/internal/http"
)

func main() {
	server, err := http.NewApp()
	if err != nil {
		slog.Error("failed to create server", "cause", err)
	}
	slog.Error("Server stopped", "errorsDuringShutdown", server.Start())
	os.Exit(1)
}
