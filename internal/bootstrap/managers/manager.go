package managers

import (
	"context"
	"database/sql"
	"log/slog"
	"maps"

	"github.com/andriihomiak/wallabago/internal/bootstrap"
	"github.com/andriihomiak/wallabago/internal/identity"
	identityStorage "github.com/andriihomiak/wallabago/internal/identity/storage"
	"github.com/andriihomiak/wallabago/internal/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Manager struct {
	identity identityStorage.SQLStorage
	boostrap storage.BoostrapSQLStorage
}

func NewManager(
	//nolint:gocritic //shadowing
	identity identityStorage.SQLStorage,
	boostrap storage.BoostrapSQLStorage,
) *Manager {
	return &Manager{
		identity: identity,
		boostrap: boostrap,
	}
}

func (m *Manager) createAdmin(ctx context.Context, appTx, identityTx *sql.Tx) error {
	// TODO: get admin username and password from config
	adminUsername := "admin"
	adminEmail := "admin@admin"
	adminPassword := "admin"

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}

	// create an identity with said username and password
	adminUser := identity.UserInfo{
		ID:           uuid.New().String(),
		Email:        adminEmail,
		Username:     adminUsername,
		PasswordHash: passwordHash,
	}

	err = m.identity.AddUserInfo(ctx, identityTx, adminUser)
	if err != nil {
		return errors.WithStack(err)
	}

	// TODO: create a user in app users with mapped identity id
	// panic("todo: implement app users saving")

	err = m.boostrap.MarkBootstrapConditionSatisfied(ctx, appTx, bootstrap.ConditionAdminCreated)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (m *Manager) createWebClient(ctx context.Context, appTx, idnetityTx *sql.Tx) error {
	client := identity.Client{
		ID:     "web",
		Secret: "web",
	}
	err := m.identity.AddClient(ctx, idnetityTx, client)
	if err != nil {
		return errors.WithStack(err)
	}
	err = m.boostrap.MarkBootstrapConditionSatisfied(ctx, appTx, bootstrap.ConditionWebClientCreated)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (m *Manager) getBootsrapSteps() map[bootstrap.ConditionName]func(context.Context, *sql.Tx, *sql.Tx) error {
	return map[bootstrap.ConditionName]func(context.Context, *sql.Tx, *sql.Tx) error{
		bootstrap.ConditionAdminCreated:     m.createAdmin,
		bootstrap.ConditionWebClientCreated: m.createWebClient,
	}
}

func (m *Manager) Bootstrap(ctx context.Context) error {
	bootstrapSteps := m.getBootsrapSteps()
	identityTx, err := m.identity.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		identityStorage.RollbackOnError(ctx, err, identityTx.Rollback)
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

	lookUp := make(map[bootstrap.ConditionName]bool)
	for _, cond := range existingConditions {
		lookUp[cond.Name] = cond.Satisfied
	}

	for conditionName, step := range maps.All(bootstrapSteps) {
		ok, satisfied := lookUp[conditionName]
		if !ok || !satisfied {
			slog.InfoContext(ctx, "Performing boostrap step", "conditionName", conditionName)
			err := step(ctx, appTx, identityTx)
			if err != nil {
				return errors.WithStack(err)
			}
			slog.InfoContext(ctx, "Boostrap step succeeded", "conditionName", conditionName)
		} else {
			slog.InfoContext(ctx, "Bootstrap condition already satisfied", "conditionName", conditionName)
		}
	}
	// todo: exclusive db locks?
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
