package app

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	stderrors "errors"

	"github.com/andriihomiak/wallabago/internal/database"
	"github.com/andriihomiak/wallabago/internal/engines"
	"github.com/andriihomiak/wallabago/internal/http/handlers"
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
}

type Wallabago struct {
	identityManager  *managers.IdentityManager
	bootstrapManager *managers.BootstrapManager
	config           *Config
	dbPool           *sql.DB
	shutdownOtel     func(context.Context) error
}

func (w *Wallabago) Addr() string {
	return w.config.Addr
}

func NewWallabago(ctx context.Context, config *Config) (*Wallabago, error) {
	// database
	dbPool, err := database.NewDBPool(ctx, config.DBConnectionString)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// storage
	bootstrapStorage := storage.NewBootstrapSQLStorage(dbPool)
	identityStorage := storage.NewIdentitySQLStorage(dbPool)
	// engines
	bootstrapEngine := engines.NewBoostrapEngine(identityStorage, bootstrapStorage)
	// managers
	boostrapManager := managers.NewBootstrapManager(bootstrapStorage, bootstrapEngine, identityStorage)
	identityManager := managers.NewIdentityManager(identityStorage)

	return &Wallabago{
		bootstrapManager: boostrapManager,
		identityManager:  identityManager,
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

	// bootstrap
	err := w.bootstrap(ctx)
	if err != nil {
		return errors.WithMessage(err, "Failed to perform bootstrap")
	}
	return nil
}

func (w *Wallabago) RegisterHandlers(mux *http.ServeMux) {
	innerMux := http.NewServeMux()

	oauth2 := handlers.NewOAuth2Handler(w.identityManager)
	innerMux.HandleFunc("POST /oauth/v2/token", oauth2.TokenEndpoint)

	ui := handlers.WebUI{}
	innerMux.HandleFunc("/", ui.Index)

	globalMiddleware := middleware.NewChain(
		middleware.NewOtelHTTPMiddleware(),
	)

	rootHandler := globalMiddleware.Wrap(innerMux)

	mux.Handle("/", rootHandler)
}

func (w *Wallabago) bootstrap(ctx context.Context) error {
	return w.bootstrapManager.Bootstrap(ctx)
}
