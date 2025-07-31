package storage

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/andriihomiak/wallabago/internal/identity"
	"github.com/pkg/errors"
)

type SQLStorage interface {
	AddClient(ctx context.Context, tx *sql.Tx, client identity.Client) error
	GetClientByID(ctx context.Context, tx *sql.Tx, id string) (*identity.Client, error)
	DeleteClientByID(ctx context.Context, tx *sql.Tx, id string) error

	AddAccessToken(ctx context.Context, tx *sql.Tx, refreshTokenID string, token identity.AccessToken) error
	GetAccessTokenByJWT(ctx context.Context, tx *sql.Tx, jwt identity.JWT) (*identity.AccessToken, error)
	RevokeAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error
	DeleteAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error

	AddRefreshToken(ctx context.Context, tx *sql.Tx, token identity.RefreshToken) error
	GetRefreshTokenByJWT(ctx context.Context, tx *sql.Tx, refreshToken identity.JWT) (*identity.RefreshToken, error)
	RevokeRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error
	DeleteRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error

	AddUserInfo(ctx context.Context, tx *sql.Tx, user identity.UserInfo) error
	GetUserInfoByUsername(ctx context.Context, tx *sql.Tx, username string) (*identity.UserInfo, error)
	DeleteUserInfoByID(ctx context.Context, tx *sql.Tx, id string) error

	Begin(ctx context.Context) (*sql.Tx, error)
}

func RollbackOnError(ctx context.Context, err error, rollbackFn func() error) {
	if err != nil {
		slog.WarnContext(ctx, "Rolling back transaction", "cause", err.Error())
		rollbackErr := rollbackFn()
		if rollbackErr != nil {
			if errors.Is(rollbackErr, sql.ErrTxDone) {
				slog.DebugContext(ctx, "Attempted to rollback already committed transaction", "cause", rollbackErr.Error())
			} else {
				slog.ErrorContext(ctx, "Error during rollback", "cause", rollbackErr.Error())
			}
		}
	}
}
