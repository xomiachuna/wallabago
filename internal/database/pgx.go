package database

import (
	"context"
	"database/sql"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx sql driver
	"github.com/pkg/errors"
)

func NewDBPool(ctx context.Context, dbURL string) (*sql.DB, error) {
	dbPool, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to setup db")
	}
	err = dbPool.PingContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to setup db")
	}
	slog.Info("Connected to db", "dbUrl", dbURL)
	return dbPool, nil
}
