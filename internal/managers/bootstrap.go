package managers

import (
	"context"
	"log/slog"
	"maps"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/engines"
	"github.com/andriihomiak/wallabago/internal/storage"
	"github.com/pkg/errors"
)

type BootstrapManager struct {
	identity storage.IdentitySQLStorage
	boostrap storage.BoostrapSQLStorage
	engine   engines.BootstrapEngine
}

func NewBootstrapManager(
	boostrap storage.BoostrapSQLStorage,
	engine engines.BootstrapEngine,
	identity storage.IdentitySQLStorage,
) *BootstrapManager {
	return &BootstrapManager{
		boostrap: boostrap,
		engine:   engine,
		identity: identity,
	}
}

func (m *BootstrapManager) getBootstrapSteps() map[core.ConditionName]func(context.Context, engines.BootstrapContext) error {
	return map[core.ConditionName]func(context.Context, engines.BootstrapContext) error{
		core.ConditionAdminCreated:     m.engine.CreateAdminAccount,
		core.ConditionWebClientCreated: m.engine.CreateWebClient,
	}
}

func (m *BootstrapManager) Bootstrap(ctx context.Context) error {
	// todo: consider db locks?
	bootstrapSteps := m.getBootstrapSteps()
	identityTx, err := m.identity.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		storage.RollbackOnError(ctx, err, identityTx.Rollback)
	}()
	appTx, err := m.boostrap.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		storage.RollbackOnError(ctx, err, appTx.Rollback)
	}()
	existingConditions, err := m.boostrap.GetBootstrapConditions(ctx, appTx)
	if err != nil {
		return errors.WithStack(err)
	}

	lookUp := make(map[core.ConditionName]bool)
	for _, cond := range existingConditions {
		lookUp[cond.Name] = cond.Satisfied
	}

	bootstrapCtx := engines.NewBoostrapContext(appTx, identityTx)

	for conditionName, step := range maps.All(bootstrapSteps) {
		ok, satisfied := lookUp[conditionName]
		if !ok || !satisfied {
			slog.InfoContext(ctx, "Performing boostrap step", "conditionName", conditionName)
			err := step(ctx, bootstrapCtx)
			if err != nil {
				return errors.WithStack(err)
			}
			slog.InfoContext(ctx, "Bootstrap step succeeded", "conditionName", conditionName)
		} else {
			slog.InfoContext(ctx, "Bootstrap condition already satisfied", "conditionName", conditionName)
		}
	}
	err = appTx.Commit()
	if err != nil {
		return errors.WithStack(err)
	}

	err = identityTx.Commit()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
