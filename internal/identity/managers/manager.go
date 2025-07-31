package managers

import (
	"context"
	"time"

	"github.com/andriihomiak/wallabago/internal/identity"
	identityStorage "github.com/andriihomiak/wallabago/internal/identity/storage"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func NewIdentityManager(
	storage identityStorage.SQLStorage,
) *IdentityManager {
	return &IdentityManager{
		storage:         storage,
		tokenExpiration: time.Hour * 24,
	}
}

type IdentityManager struct {
	storage         identityStorage.SQLStorage
	key             []byte
	tokenExpiration time.Duration
}

func (m *IdentityManager) PasswordFlow(ctx context.Context, req identity.PasswordFlowRequest) (*identity.AccessTokenResponse, error) {
	tx, err := m.storage.Begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		identityStorage.RollbackOnError(ctx, err, tx.Rollback)
	}()

	// check client credentials
	client, err := m.storage.GetClientByID(ctx, tx, req.ClientID)
	if err != nil {
		// todo: check error type
		return nil, &identity.AuthError{
			ErrorName:        identity.AuthErrorInvalidClient,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}
	if client.Secret != req.ClientSecret {
		return nil, &identity.AuthError{
			ErrorName:        identity.AuthErrorInvalidClient,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}

	// check user credentials
	user, err := m.storage.GetUserInfoByUsername(ctx, tx, req.Username)
	if err != nil {
		// todo: check error type
		return nil, &identity.AuthError{
			ErrorName:        identity.AuthErrorInvalidGrant,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}

	passwordErr := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(req.Password))
	if passwordErr != nil {
		// todo: check error type
		return nil, &identity.AuthError{
			ErrorName:        identity.AuthErrorInvalidGrant,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}

	var scope *identity.Scope
	if req.Scope == "" {
		scope = identity.DefaultScope()
	} else {
		scope, err = identity.NewScopeFromString(req.Scope)
		if err != nil {
			return nil, &identity.AuthError{
				ErrorName:        identity.AuthErrorInvalidScope,
				ErrorDescription: errors.WithStack(err).Error(),
			}
		}
	}

	// credentials correct at this point, issue a new token pair

	// create and save refresh token
	refreshToken, err := identity.NewRefreshToken(user.ID, client.ID, m.key)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = m.storage.AddRefreshToken(ctx, tx, *refreshToken)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// create and save access token
	accessStoken, err := identity.NewAccessToken(user.ID, client.ID, *scope, m.tokenExpiration, m.key)
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

	return &identity.AccessTokenResponse{
		AccessToken:  *accessStoken,
		RefreshToken: refreshToken.Token,
	}, nil
}

func (m *IdentityManager) RefreshTokenFlow(ctx context.Context, req identity.RefreshTokenFlowRequest) (*identity.AccessTokenResponse, error) {
	tx, err := m.storage.Begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		identityStorage.RollbackOnError(ctx, err, tx.Rollback)
	}()

	// check client credentials
	client, err := m.storage.GetClientByID(ctx, tx, req.ClientID)
	if err != nil {
		// todo: check error type
		return nil, &identity.AuthError{
			ErrorName:        identity.AuthErrorInvalidClient,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}
	if client.Secret != req.ClientSecret {
		return nil, &identity.AuthError{
			ErrorName:        identity.AuthErrorInvalidClient,
			ErrorDescription: errors.WithStack(err).Error(),
		}
	}
	// revoke previous access tokens of this refresh token
	// issue a new token
	// save the token
	// return
	panic("todo")
}
