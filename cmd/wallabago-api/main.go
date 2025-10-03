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

	dbPath := os.Getenv("WALLABAGO_DB_PATH")
	if dbPath == "" {
		dbPath = "./wallabago.db"
	}

	bootstrapAdminUsername := os.Getenv("WALLABAGO_BOOTSTRAP_ADMIN_USERNAME")
	if bootstrapAdminUsername == "" {
		bootstrapAdminUsername = "admin"
	}

	bootstrapAdminPassword := os.Getenv("WALLABAGO_BOOTSTRAP_ADMIN_PASSWORD")
	if bootstrapAdminPassword == "" {
		bootstrapAdminPassword = "admin"
	}

	bootstrapAdminEmail := os.Getenv("WALLABAGO_BOOTSTRAP_ADMIN_EMAIL")
	if bootstrapAdminEmail == "" {
		bootstrapAdminEmail = "admin@admin.co"
	}

	bootstrapClientID := os.Getenv("WALLABAGO_BOOTSTRAP_CLIENT_ID")
	if bootstrapClientID == "" {
		bootstrapClientID = "web"
	}

	bootstrapClientSecret := os.Getenv("WALLABAGO_BOOTSTRAP_CLIENT_SECRET")
	if bootstrapClientSecret == "" {
		bootstrapClientSecret = "web"
	}

	server, err := http.NewServer(
		context.TODO(),
		app.Config{
			Addr:                   addr,
			InstrumentationEnabled: instrument,
			DBPath:                 dbPath,
			BootstrapAdminUsername: bootstrapAdminUsername,
			BootstrapAdminPassword: bootstrapAdminPassword,
			BootstrapAdminEmail:    bootstrapAdminEmail,
			BootstrapClientID:      bootstrapClientID,
			BootstrapClientSecret:  bootstrapClientSecret,
		},
	)
	if err != nil {
		slog.Error("failed to create server", "cause", err)
	}
	slog.Error("Server stopped", "errorsDuringShutdown", server.Start(context.Background()))
	os.Exit(1)
}
