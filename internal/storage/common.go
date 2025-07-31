package storage

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/pkg/errors"
)

type TransactionStarter interface {
	Begin(ctx context.Context) (*sql.Tx, error)
}

func RollbackOnError(ctx context.Context, err error, rollbackFn func() error) {
	if err != nil {
		slog.WarnContext(ctx, "Rolling back transaction", "cause", err.Error())
		rollbackErr := rollbackFn()
		if rollbackErr != nil {
			if errors.Is(rollbackErr, sql.ErrTxDone) {
				slog.DebugContext(ctx, "Attempted to rollback already committed transaction", "cause", rollbackErr.Error())
			} else {
				slog.ErrorContext(ctx, "Error during rollback", "cause", rollbackErr.Error())
			}
		}
	}
}
