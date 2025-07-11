package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"go.opentelemetry.io/otel/codes"
)

func writeJSON[T any](ctx context.Context, w http.ResponseWriter, value T, status int) error {
	ctx, span := tracer.Start(ctx, "write-json")
	defer span.End()
	body, err := json.Marshal(value)
	if err != nil {
        panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(body)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to write JSON body", "body", body, "error", err)
	} else {
		logger.DebugContext(ctx, "Sending JSON Response", "body", body)
	}
	return err
}

type indexResponse struct {
    Page string `json:"page"`
    Data string `json:"data,omitempty"`
}

func setupServer(baseCtx context.Context) *http.Server {
    middleware := setupMiddleware()
	ctx, span := tracer.Start(baseCtx, "handler-setup")
	defer span.End()
	mux := http.DefaultServeMux

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) { // exact match
		logger.InfoContext(r.Context(), "Handling request", "page", r.URL, "method", r.Method)
		// time.Sleep(1000 * time.Millisecond)
		writeJSON(r.Context(), w, indexResponse{Page: r.URL.String()}, http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { // anything else
		slog.Warn("Page not found", "page", r.URL)
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /protected", func(w http.ResponseWriter, r *http.Request) { // anything else
		ctx, span := tracer.Start(r.Context(), "auth-check")
		defer span.End()
		logger.InfoContext(ctx, "Handling request", "page", r.URL, "method", r.Method)
		var engine AuthenticationEngine = &HardcodedAuthnEngine{
			username: "admin",
			password: "password",
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			span.SetStatus(codes.Error, "missing_auth_header")
			span.AddEvent("initiate_basic_auth")
			w.Header().Set("www-authenticate", "Basic")
			writeJSON(ctx, w, &map[string]string{"error": "unauthorized"}, http.StatusUnauthorized)
			logger.InfoContext(ctx, "Unauthorized", "cause", "missing_auth_header")
		} else if strings.HasPrefix(authHeader, "Basic ") {
			token, _ := strings.CutPrefix(authHeader, "Basic ")
			user, err := engine.BasicAuthn(ctx, BasicAuthnToken(token))
			if err != nil {
				w.Header().Set("www-authenticate", "Basic")
				writeJSON(ctx, w, &map[string]string{"error": "unauthorized"}, http.StatusUnauthorized)
				logger.InfoContext(ctx, "Unauthorized", "cause", "authn_error", "detail", err)
				span.SetStatus(codes.Error, err.Error())
				return
			}
			span.End()
			// time.Sleep(time.Millisecond * 100)
			writeJSON(ctx, w, user, http.StatusOK)
			logger.InfoContext(ctx, "Visited protected page", "user", user, "page", r.URL)
			span.SetStatus(http.StatusOK, "")
		} else {
			w.Header().Set("www-authenticate", "Basic")
			writeJSON(ctx, w, &map[string]string{"error": "unauthorized"}, http.StatusUnauthorized)
			logger.InfoContext(ctx, "Unauthorized", "cause", "missing_auth_header")
			span.SetStatus(codes.Error, "missing basic auth header")
		}
	})
	server := &http.Server{
		Handler:     middleware.wrap(mux),
		Addr:        ":9999",
		BaseContext: func(_ net.Listener) context.Context { return baseCtx },
	}
	logger.InfoContext(ctx, "Created server", "addr", server.Addr)
	return server
}

func main() {
	ctx, stopSignalHandler := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignalHandler()
	shutdown, err := setupOtelSdk(ctx)
	if err != nil {
		log.Fatalf("failed to setup otel: %w", err)
	}
	defer func() {
		// print all the errors and exit
		err = errors.Join(shutdown(ctx))
		if err != nil {
			log.Fatal(err)
		}
	}()

	server := setupServer(ctx)
	serverErr := make(chan error, 1)
	go func(serverErr chan<- error) {
		serverErr <- server.ListenAndServe()
	}(serverErr)
	select {
	case err = <-serverErr:
		return
	case <-ctx.Done():
		stopSignalHandler()
		server.Shutdown(context.Background())
	}
}
