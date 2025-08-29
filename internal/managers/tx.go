package managers

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/pkg/errors"
)

type transactionStarter interface {
	// we assume that we have a single database so implementations that have this
	// are all compatible and interchangeable with regards to starting the transaction
	Begin(ctx context.Context) (*sql.Tx, error)
}

func rollbackOnError(ctx context.Context, err error, rollbackFn func() error) {
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
