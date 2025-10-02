package engines

import (
	"context"
	"database/sql"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/core/policy"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type BootstrapStorage interface {
	AddClient(ctx context.Context, tx *sql.Tx, client core.Client) error
	AddUserInfo(ctx context.Context, tx *sql.Tx, user core.UserInfo) error
	AddUser(ctx context.Context, tx *sql.Tx, user core.User) error
	AssignUserRole(ctx context.Context, tx *sql.Tx, userID string, role policy.Role) error
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

func (e *BootstrapEngine) CreateAdminAccount(ctx context.Context, tx *sql.Tx, admin core.BootstrapAdminCredentials) error {
	// Note: This assumes only the bootstrapped admin user is created via this flow.
	// Regular user creation with role assignment will be handled separately.

	// TODO: get admin username and password from config?
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}

	// create an identity with said username and password
	adminUser := core.UserInfo{
		ID:           uuid.New().String(),
		Email:        admin.Email,
		Username:     admin.Username,
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

	// Assign admin role for RBAC
	err = e.storage.AssignUserRole(ctx, tx, adminUser.ID, policy.RoleAdmin)
	if err != nil {
		return errors.WithStack(err)
	}

	err = e.storage.MarkBootstrapConditionSatisfied(ctx, tx, core.ConditionAdminCreated)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (e *BootstrapEngine) CreateInitialClient(ctx context.Context, tx *sql.Tx, client core.Client) error {
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
