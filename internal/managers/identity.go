package managers

import (
	"context"
	"database/sql"
	"time"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type IdentityStorage interface {
	AddClient(ctx context.Context, tx *sql.Tx, client core.Client) error
	GetClientByID(ctx context.Context, tx *sql.Tx, id string) (*core.Client, error)
	DeleteClientByID(ctx context.Context, tx *sql.Tx, id string) error

	AddAccessToken(ctx context.Context, tx *sql.Tx, refreshTokenID string, token core.AccessToken) error
	GetAccessTokenByJWT(ctx context.Context, tx *sql.Tx, jwt core.JWT) (*core.AccessToken, error)
	RevokeAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error
	DeleteAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error

	AddRefreshToken(ctx context.Context, tx *sql.Tx, token core.RefreshToken) error
	GetRefreshTokenByJWT(ctx context.Context, tx *sql.Tx, refreshToken core.JWT) (*core.RefreshToken, error)
	RevokeRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error
	DeleteRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error

	AddUserInfo(ctx context.Context, tx *sql.Tx, user core.UserInfo) error
	GetUserInfoByUsername(ctx context.Context, tx *sql.Tx, username string) (*core.UserInfo, error)
	DeleteUserInfoByID(ctx context.Context, tx *sql.Tx, id string) error

	transactionStarter
}

func NewIdentityManager(
	identityStorage IdentityStorage,
) *IdentityManager {
	return &IdentityManager{
		storage:         identityStorage,
		tokenExpiration: time.Hour * 24,
		key:             []byte("wallabago"), // todo: pass from outside
	}
}

type IdentityManager struct {
	storage         IdentityStorage
	key             []byte
	tokenExpiration time.Duration
}

func (m *IdentityManager) PasswordFlow(ctx context.Context, req core.PasswordFlowRequest) (*core.AccessTokenResponse, error) {
	tx, err := m.storage.Begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		rollbackOnError(ctx, err, tx.Rollback)
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
			ErrorDescription: "Bad client credentials",
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
			ErrorDescription: errors.WithStack(passwordErr).Error(),
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
		rollbackOnError(ctx, err, tx.Rollback)
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

func (m *IdentityManager) Authenticate(ctx context.Context, accessToken string) (*core.AccessToken, error) {
	// todo: check signature first
	tx, err := m.storage.Begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	//nolint:errcheck //ok here since we dont do any writes
	defer tx.Rollback()
	token, err := m.storage.GetAccessTokenByJWT(ctx, tx, core.JWT(accessToken))
	if err != nil {
		return nil, &core.AuthError{
			ErrorName: core.AuthErrorUnauthorized,
		}
	}
	if token.Revoked {
		return nil, &core.AuthError{
			ErrorName: core.AuthErrorUnauthorized,
		}
	}
	return token, nil
}
