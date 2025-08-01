package managers

import (
	"context"
	"time"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/storage"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func NewIdentityManager(
	identityStorage storage.IdentitySQLStorage,
) *IdentityManager {
	return &IdentityManager{
		storage:         identityStorage,
		tokenExpiration: time.Hour * 24,
	}
}

type IdentityManager struct {
	storage         storage.IdentitySQLStorage
	key             []byte
	tokenExpiration time.Duration
}

func (m *IdentityManager) PasswordFlow(ctx context.Context, req core.PasswordFlowRequest) (*core.AccessTokenResponse, error) {
	tx, err := m.storage.Begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		storage.RollbackOnError(ctx, err, tx.Rollback)
	}()

	// check client credentials
	client, err := m.storage.GetClientByID(ctx, tx, req.ClientID)
	if err != nil {
		// todo: check error type
		return nil, &core.AuthError{
			ErrorName:        core.AuthErrorInvalidClient,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}
	if client.Secret != req.ClientSecret {
		return nil, &core.AuthError{
			ErrorName:        core.AuthErrorInvalidClient,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}

	// check user credentials
	user, err := m.storage.GetUserInfoByUsername(ctx, tx, req.Username)
	if err != nil {
		// todo: check error type
		return nil, &core.AuthError{
			ErrorName:        core.AuthErrorInvalidGrant,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}

	passwordErr := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(req.Password))
	if passwordErr != nil {
		// todo: check error type
		return nil, &core.AuthError{
			ErrorName:        core.AuthErrorInvalidGrant,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}

	var scope *core.Scope
	if req.Scope == "" {
		scope = core.DefaultScope()
	} else {
		scope, err = core.NewScopeFromString(req.Scope)
		if err != nil {
			return nil, &core.AuthError{
				ErrorName:        core.AuthErrorInvalidScope,
				ErrorDescription: errors.WithStack(err).Error(),
			}
		}
	}

	// credentials correct at this point, issue a new token pair

	// create and save refresh token
	refreshToken, err := core.NewRefreshToken(user.ID, client.ID, m.key)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = m.storage.AddRefreshToken(ctx, tx, *refreshToken)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// create and save access token
	accessStoken, err := core.NewAccessToken(user.ID, client.ID, *scope, m.tokenExpiration, m.key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = m.storage.AddAccessToken(ctx, tx, refreshToken.ID, *accessStoken)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &core.AccessTokenResponse{
		AccessToken:  *accessStoken,
		RefreshToken: refreshToken.Token,
	}, nil
}

func (m *IdentityManager) RefreshTokenFlow(ctx context.Context, req core.RefreshTokenFlowRequest) (*core.AccessTokenResponse, error) {
	tx, err := m.storage.Begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		storage.RollbackOnError(ctx, err, tx.Rollback)
	}()

	// check client credentials
	client, err := m.storage.GetClientByID(ctx, tx, req.ClientID)
	if err != nil {
		// todo: check error type
		return nil, &core.AuthError{
			ErrorName:        core.AuthErrorInvalidClient,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}
	if client.Secret != req.ClientSecret {
		return nil, &core.AuthError{
			ErrorName:        core.AuthErrorInvalidClient,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}
	// revoke previous access tokens of this refresh token
	// issue a new token
	// save the token
	// return
	panic("todo")
}
