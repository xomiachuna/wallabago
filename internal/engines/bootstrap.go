package engines

import (
	"context"
	"database/sql"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/storage"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type BootstrapContext struct {
	appTx      *sql.Tx
	identityTx *sql.Tx
}

func NewBoostrapContext(appTx, identityTx *sql.Tx) BootstrapContext {
	return BootstrapContext{
		appTx:      appTx,
		identityTx: identityTx,
	}
}

type BootstrapEngine interface {
	CreateAdminAccount(context.Context, BootstrapContext) error
	CreateWebClient(context.Context, BootstrapContext) error
}

type bootstrapEngine struct {
	identity storage.IdentitySQLStorage
	boostrap storage.BoostrapSQLStorage
}

func NewBoostrapEngine(
	identity storage.IdentitySQLStorage,
	boostrap storage.BoostrapSQLStorage,
) BootstrapEngine {
	return &bootstrapEngine{
		identity: identity,
		boostrap: boostrap,
	}
}

var _ BootstrapEngine = (*bootstrapEngine)(nil)

func (e *bootstrapEngine) CreateAdminAccount(ctx context.Context, bootstrapCtx BootstrapContext) error {
	// TODO: get admin username and password from config
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

	err = e.identity.AddUserInfo(ctx, bootstrapCtx.identityTx, adminUser)
	if err != nil {
		return errors.WithStack(err)
	}

	// TODO: create a user in app users with mapped identity id
	// panic("todo: implement app users saving")

	err = e.boostrap.MarkBootstrapConditionSatisfied(ctx, bootstrapCtx.appTx, core.ConditionAdminCreated)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (e *bootstrapEngine) CreateWebClient(ctx context.Context, bootstrapCtx BootstrapContext) error {
	client := core.Client{
		ID:     "web",
		Secret: "web",
	}
	err := e.identity.AddClient(ctx, bootstrapCtx.identityTx, client)
	if err != nil {
		return errors.WithStack(err)
	}
	err = e.boostrap.MarkBootstrapConditionSatisfied(ctx, bootstrapCtx.appTx, core.ConditionWebClientCreated)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
