package storage

import (
	"context"
	"database/sql"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/database"
	"github.com/pkg/errors"
)

type IdentitySQLStorage interface {
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

	TransactionStarter
}

// identitySQLStorage is an implementation of [IdentitySQLStorage]
// that uses sqlc codegen to access the database.
type identitySQLStorage struct {
	queries *database.Queries
	db      *sql.DB
}

func NewIdentitySQLStorage(db *sql.DB) IdentitySQLStorage {
	return &identitySQLStorage{
		queries: database.New(db),
		db:      db,
	}
}

var _ IdentitySQLStorage = (*identitySQLStorage)(nil)

func (s *identitySQLStorage) Begin(ctx context.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, nil)
}

func (s *identitySQLStorage) AddClient(ctx context.Context, tx *sql.Tx, client core.Client) error {
	q := s.queries.WithTx(tx)
	_, err := q.AddClient(ctx, database.AddClientParams{
		ClientID:     client.ID,
		ClientSecret: client.Secret,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *identitySQLStorage) GetClientByID(ctx context.Context, tx *sql.Tx, id string) (*core.Client, error) {
	q := s.queries.WithTx(tx)
	result, err := q.GetClientByID(ctx, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &core.Client{
		ID:     result.ClientID,
		Secret: result.ClientSecret,
	}, nil
}

func (s *identitySQLStorage) DeleteClientByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteClientByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *identitySQLStorage) AddAccessToken(ctx context.Context, tx *sql.Tx, refreshTokenID string, token core.AccessToken) error {
	q := s.queries.WithTx(tx)
	_, err := q.AddAccessToken(ctx, database.AddAccessTokenParams{
		TokenID:  token.ID,
		ClientID: token.ClientID,
		Jwt:      string(token.Token),
		UserID:   token.UserID,
		Revoked:  token.Revoked,
		RefreshTokenID: sql.NullString{
			Valid:  refreshTokenID != "",
			String: refreshTokenID,
		},
		Type:             string(token.TokenType),
		Scope:            string(token.Scope),
		IssuedAt:         token.IssuedAt,
		ExpiresInSeconds: token.ExpiresInSeconds,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *identitySQLStorage) GetAccessTokenByJWT(ctx context.Context, tx *sql.Tx, jwt core.JWT) (*core.AccessToken, error) {
	q := s.queries.WithTx(tx)
	result, err := q.GetAccessTokenByJWT(ctx, string(jwt))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &core.AccessToken{
		ID:               result.ClientID,
		Token:            core.JWT(result.Jwt),
		ExpiresInSeconds: result.ExpiresInSeconds,
		UserID:           result.UserID,
		Scope:            core.Scope(result.Scope),
		IssuedAt:         result.IssuedAt,
		TokenType:        core.TokenType(result.Type),
		ClientID:         result.ClientID,
		Revoked:          result.Revoked,
	}, nil
}

func (s *identitySQLStorage) RevokeAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	_, err := q.RevokeAccessTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *identitySQLStorage) DeleteAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteAccessTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *identitySQLStorage) AddRefreshToken(ctx context.Context, tx *sql.Tx, token core.RefreshToken) error {
	q := s.queries.WithTx(tx)
	_, err := q.AddRefreshToken(ctx, database.AddRefreshTokenParams{
		TokenID:  token.ID,
		Jwt:      string(token.Token),
		ClientID: token.ClientID,
		Revoked:  token.Revoked,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *identitySQLStorage) GetRefreshTokenByJWT(ctx context.Context, tx *sql.Tx, refreshToken core.JWT) (*core.RefreshToken, error) {
	q := s.queries.WithTx(tx)
	result, err := q.GetRefreshTokenByJWT(ctx, string(refreshToken))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &core.RefreshToken{
		ID:       result.ClientID,
		Token:    core.JWT(result.Jwt),
		ClientID: result.ClientID,
		Revoked:  result.Revoked,
	}, nil
}

func (s *identitySQLStorage) RevokeRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	_, err := q.RevokeRefreshTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *identitySQLStorage) DeleteRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteRefreshTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *identitySQLStorage) AddUserInfo(ctx context.Context, tx *sql.Tx, user core.UserInfo) error {
	q := s.queries.WithTx(tx)
	_, err := q.AddIdentityUser(ctx, database.AddIdentityUserParams{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *identitySQLStorage) GetUserInfoByUsername(ctx context.Context, tx *sql.Tx, username string) (*core.UserInfo, error) {
	q := s.queries.WithTx(tx)
	result, err := q.GetIdentityUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &core.UserInfo{
		ID:           result.UserID,
		Email:        result.Email,
		Username:     result.Username,
		PasswordHash: result.PasswordHash,
	}, nil
}

func (s *identitySQLStorage) DeleteUserInfoByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteIdentityUserByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
