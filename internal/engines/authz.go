package engines

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/andriihomiak/wallabago/internal/core/policy"
)

type NoopAuthorizationEngine struct {
	userStorage RBACUserStorage
}

func NewRBACAuthorizationEngine(
	userStorage RBACUserStorage,
) *NoopAuthorizationEngine {
	return &NoopAuthorizationEngine{
		userStorage: userStorage,
	}
}

func (e *NoopAuthorizationEngine) Begin(ctx context.Context) (*sql.Tx, error) {
	return e.userStorage.Begin(ctx)
}

func (e *NoopAuthorizationEngine) CheckPolicy(ctx context.Context, _ *sql.Tx, userID string, action policy.Action) error {
	slog.WarnContext(ctx, "noop policy check", "userID", userID, "action", action)
	return nil
}

type RBACUserStorage interface {
	Begin(ctx context.Context) (*sql.Tx, error)
}
