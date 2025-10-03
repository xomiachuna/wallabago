package database

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/pkg/errors"
	_ "modernc.org/sqlite" // SQLite driver
)

func NewDBPool(ctx context.Context, dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Set SQLite pragmas for optimal performance and safety
	pragmas := []string{
		"PRAGMA journal_mode = WAL",   // Write-Ahead Logging for better concurrency
		"PRAGMA synchronous = NORMAL", // Good balance of safety and performance
		"PRAGMA foreign_keys = ON",    // Enable foreign key constraints
		"PRAGMA busy_timeout = 5000",  // Wait up to 5s when database is locked
		"PRAGMA cache_size = -64000",  // 64MB cache (negative = KB, positive = pages)
		"PRAGMA temp_store = MEMORY",  // Store temp tables in memory
	}

	for _, pragma := range pragmas {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			return nil, errors.Wrapf(err, "failed to set pragma: %s", pragma)
		}
	}

	// Run migrations
	if err := MigrateSQLite(db); err != nil {
		return nil, errors.Wrap(err, "failed to run migrations")
	}

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	slog.InfoContext(ctx, "Connected to SQLite database", "path", dbPath)
	return db, nil
}
