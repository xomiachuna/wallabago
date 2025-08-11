package managers

import (
	"context"
	"database/sql"
	"log/slog"
	"maps"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/pkg/errors"
)

type BootstrapStorage interface {
	GetBootstrapConditions(ctx context.Context, tx *sql.Tx) ([]core.Condition, error)

	transactionStarter
}

type BootstrapEngine interface {
	CreateWebClient(context.Context, *sql.Tx) error
	CreateAdminAccount(context.Context, *sql.Tx) error
}

type BootstrapManager struct {
	storage BootstrapStorage
	engine  BootstrapEngine
}

func NewBootstrapManager(
	storage BootstrapStorage,
	engine BootstrapEngine,
) *BootstrapManager {
	return &BootstrapManager{
		storage: storage,
		engine:  engine,
	}
}

func (m *BootstrapManager) getBootstrapSteps() map[core.ConditionName]func(context.Context, *sql.Tx) error {
	return map[core.ConditionName]func(context.Context, *sql.Tx) error{
		core.ConditionAdminCreated:     m.engine.CreateAdminAccount,
		core.ConditionWebClientCreated: m.engine.CreateWebClient,
	}
}

func (m *BootstrapManager) Bootstrap(ctx context.Context) error {
	// todo: consider db locks?
	bootstrapSteps := m.getBootstrapSteps()
	tx, err := m.storage.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		rollbackOnError(ctx, err, tx.Rollback)
	}()
	existingConditions, err := m.storage.GetBootstrapConditions(ctx, tx)
	if err != nil {
		return errors.WithStack(err)
	}

	lookUp := make(map[core.ConditionName]bool)
	for _, cond := range existingConditions {
		lookUp[cond.Name] = cond.Satisfied
	}

	for conditionName, step := range maps.All(bootstrapSteps) {
		ok, satisfied := lookUp[conditionName]
		if !ok || !satisfied {
			slog.InfoContext(ctx, "Performing boostrap step", "conditionName", conditionName)
			err := step(ctx, tx)
			if err != nil {
				return errors.WithStack(err)
			}
			slog.InfoContext(ctx, "Bootstrap step succeeded", "conditionName", conditionName)
		} else {
			slog.InfoContext(ctx, "Bootstrap condition already satisfied", "conditionName", conditionName)
		}
	}
	err = tx.Commit()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
