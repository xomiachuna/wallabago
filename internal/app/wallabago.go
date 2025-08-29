package app

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	stderrors "errors"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/database"
	"github.com/andriihomiak/wallabago/internal/engines"
	"github.com/andriihomiak/wallabago/internal/http/handlers"
	"github.com/andriihomiak/wallabago/internal/http/handlers/docs"
	"github.com/andriihomiak/wallabago/internal/http/middleware"
	"github.com/andriihomiak/wallabago/internal/instrumentation"
	"github.com/andriihomiak/wallabago/internal/managers"
	"github.com/andriihomiak/wallabago/internal/storage"
	"github.com/pkg/errors"
)

type Config struct {
	Addr                   string
	InstrumentationEnabled bool
	DBConnectionString     string

	BootstrapAdminEmail, BootstrapAdminUsername, BootstrapAdminPassword string
	BootstrapClientID, BootstrapClientSecret                            string
}

type Wallabago struct {
	identityManager  *managers.IdentityManager
	bootstrapManager *managers.BootstrapManager
	entryManager     *managers.EntryManager
	config           *Config
	dbPool           *sql.DB
	shutdownOtel     func(context.Context) error
}

func (w *Wallabago) Addr() string {
	return w.config.Addr
}

func (w *Wallabago) Config() *Config {
	return w.config
}

func NewWallabago(ctx context.Context, config *Config) (*Wallabago, error) {
	// database
	dbPool, err := database.NewDBPool(ctx, config.DBConnectionString)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	postgresStorage := storage.NewPostreSQLStorage(dbPool)
	// engines
	bootstrapEngine := engines.NewBoostrapEngine(postgresStorage)
	// managers
	boostrapManager := managers.NewBootstrapManager(postgresStorage, bootstrapEngine, core.BootstrapAdminCredentials{
		Username: config.BootstrapAdminUsername,
		Password: config.BootstrapAdminPassword,
		Email:    config.BootstrapAdminEmail,
	}, core.Client{
		ID:     config.BootstrapClientID,
		Secret: config.BootstrapClientSecret,
	})
	identityManager := managers.NewIdentityManager(postgresStorage)

	entryManager := managers.NewEntryManager(nil, nil, nil)

	return &Wallabago{
		bootstrapManager: boostrapManager,
		identityManager:  identityManager,
		entryManager:     entryManager,
		config:           config,
		dbPool:           dbPool,
		shutdownOtel: func(ctx context.Context) error {
			slog.WarnContext(ctx, "Otel instrumentation is not enabled, nothing to cleanup"+
				"In order to enable instrumentation pass config.InstrumentationEnabled and use Wallabago.Prepare()")
			return nil
		},
	}, nil
}

func (w *Wallabago) shutdownDB(shutdownCtx context.Context) error {
	slog.InfoContext(shutdownCtx, "Closing db pool")
	err := w.dbPool.Close()
	if err != nil {
		slog.WarnContext(shutdownCtx, "Error occurred during db pool shutdown", "err", err.Error())
		return errors.WithStack(err)
	}

	slog.InfoContext(shutdownCtx, "Db pool closed")
	return nil
}

func (w *Wallabago) Shutdown(shutdownCtx context.Context) error {
	slog.Warn("Wallabago is shutting down")
	otelShutdownErr := errors.Wrap(w.shutdownOtel(shutdownCtx), "errors during otel shutdown")
	dbShutdownErr := errors.Wrap(w.shutdownDB(shutdownCtx), "errors during db shutdown")
	return stderrors.Join(otelShutdownErr, dbShutdownErr)
}

func (w *Wallabago) Prepare(ctx context.Context) error {
	// otel
	if w.config.InstrumentationEnabled {
		shutdownOtel, err := instrumentation.SetupOtelSDK(ctx)
		if err != nil {
			return errors.Wrap(err, "Failed to setup otel")
		}
		w.shutdownOtel = shutdownOtel
	}

	// run bootstrap
	err := w.bootstrap(ctx)
	if err != nil {
		return errors.WithMessage(err, "Failed to perform bootstrap")
	}
	return nil
}

func (w *Wallabago) Handler() http.Handler {
	mux := http.NewServeMux()

	auth := middleware.NewOAuth2Middleware(w.identityManager)

	oauth2 := handlers.NewOAuth2Handler(w.identityManager)
	mux.HandleFunc("POST /oauth/v2/token", oauth2.TokenEndpoint)

	ui := handlers.NewWebUI()
	api := handlers.NewAPI(
		w.entryManager,
	)

	mux.HandleFunc("/{$}", ui.Index)
	mux.Handle("/docs/", http.StripPrefix("/docs/", docs.OpenAPI))
	mux.Handle("GET /api/entries/exists", middleware.WrapFunc(api.EntryExists, auth))
	mux.Handle("POST /api/entries", middleware.WrapFunc(api.AddEntry, auth))

	globalMiddleware := middleware.NewChain(
		middleware.LoggingMiddleware,
		middleware.NewOtelHTTPMiddleware(),
	)

	return globalMiddleware.Wrap(mux)
}

func (w *Wallabago) bootstrap(ctx context.Context) error {
	return w.bootstrapManager.Bootstrap(ctx)
}
