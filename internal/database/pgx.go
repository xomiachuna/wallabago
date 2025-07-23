package database

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxstdlib "github.com/jackc/pgx/v5/stdlib" // pgx sql driver
	"github.com/pkg/errors"
)

func NewDBPool(ctx context.Context, dbURL string) (*sql.DB, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dbPool := pgxstdlib.OpenDBFromPool(pool)
	err = dbPool.PingContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to setup db")
	}
	slog.Info("Connected to db", "dbUrl", dbURL)
	return dbPool, nil
}
