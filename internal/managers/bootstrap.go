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
	CreateInitialClient(context.Context, *sql.Tx, core.Client) error
	CreateAdminAccount(context.Context, *sql.Tx, core.BootstrapAdminCredentials) error
}

type BootstrapManager struct {
	storage BootstrapStorage
	engine  BootstrapEngine

	bootstrapAdminCredentials core.BootstrapAdminCredentials
	bootstrapClient           core.Client
}

func NewBootstrapManager(
	storage BootstrapStorage,
	engine BootstrapEngine,
	bootstrapAdminCredentials core.BootstrapAdminCredentials,
	bootstrapClient core.Client,
) *BootstrapManager {
	return &BootstrapManager{
		storage:                   storage,
		engine:                    engine,
		bootstrapAdminCredentials: bootstrapAdminCredentials,
		bootstrapClient:           bootstrapClient,
	}
}

func (m *BootstrapManager) getBootstrapSteps() map[core.ConditionName]func(context.Context, *sql.Tx) error {
	return map[core.ConditionName]func(context.Context, *sql.Tx) error{
		core.ConditionAdminCreated: func(ctx context.Context, tx *sql.Tx) error {
			return m.engine.CreateAdminAccount(ctx, tx, m.bootstrapAdminCredentials)
		},
		core.ConditionWebClientCreated: func(ctx context.Context, tx *sql.Tx) error {
			return m.engine.CreateInitialClient(ctx, tx, m.bootstrapClient)
		},
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
