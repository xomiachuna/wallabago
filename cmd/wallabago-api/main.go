package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/andriihomiak/wallabago/internal/app"
	"github.com/andriihomiak/wallabago/internal/http"
)

func main() {
	addr := "0.0.0.0:8080"
	if port, ok := os.LookupEnv("WALLABAGO_PORT"); ok {
		// todo: check if port is int?
		addr = fmt.Sprintf("0.0.0.0:%s", port)
	}
	_, instrument := os.LookupEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	dbConnString := os.Getenv("DB")
	server, err := http.NewServer(
		context.TODO(),
		app.Config{
			Addr:                   addr,
			InstrumentationEnabled: instrument,
			DBConnectionString:     dbConnString,
		},
	)
	if err != nil {
		slog.Error("failed to create server", "cause", err)
	}
	slog.Error("Server stopped", "errorsDuringShutdown", server.Start(context.Background()))
	os.Exit(1)
}
