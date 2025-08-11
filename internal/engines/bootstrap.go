package engines

import (
	"context"
	"database/sql"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type BootstrapStorage interface {
	AddClient(ctx context.Context, tx *sql.Tx, client core.Client) error
	AddUserInfo(ctx context.Context, tx *sql.Tx, user core.UserInfo) error
	AddUser(ctx context.Context, tx *sql.Tx, user core.User) error
	GetBootstrapConditions(ctx context.Context, tx *sql.Tx) ([]core.Condition, error)
	MarkBootstrapConditionSatisfied(ctx context.Context, tx *sql.Tx, condition core.ConditionName) error
}

type BootstrapEngine struct {
	storage BootstrapStorage
}

func NewBoostrapEngine(
	storage BootstrapStorage,
) *BootstrapEngine {
	return &BootstrapEngine{
		storage: storage,
	}
}

func (e *BootstrapEngine) CreateAdminAccount(ctx context.Context, tx *sql.Tx) error {
	// TODO: get admin username and password from config?
	adminUsername := "admin"
	adminEmail := "admin@admin"
	adminPassword := "admin"

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}

	// create an identity with said username and password
	adminUser := core.UserInfo{
		ID:           uuid.New().String(),
		Email:        adminEmail,
		Username:     adminUsername,
		PasswordHash: passwordHash,
	}

	err = e.storage.AddUserInfo(ctx, tx, adminUser)
	if err != nil {
		return errors.WithStack(err)
	}

	err = e.storage.AddUser(ctx, tx, core.User{
		ID:       adminUser.ID,
		IsAdmin:  true,
		Username: adminUser.Username,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	err = e.storage.MarkBootstrapConditionSatisfied(ctx, tx, core.ConditionAdminCreated)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (e *BootstrapEngine) CreateWebClient(ctx context.Context, tx *sql.Tx) error {
	client := core.Client{
		ID:     "web",
		Secret: "web",
	}
	err := e.storage.AddClient(ctx, tx, client)
	if err != nil {
		return errors.WithStack(err)
	}
	err = e.storage.MarkBootstrapConditionSatisfied(ctx, tx, core.ConditionWebClientCreated)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
