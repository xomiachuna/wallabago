package storage

import (
	"context"
	"database/sql"

	"github.com/andriihomiak/wallabago/internal/identity"
	"github.com/andriihomiak/wallabago/internal/identity/database"
	"github.com/pkg/errors"
)

// CodegenStorage is an implementation of [SQLStorage]
// that uses sqlc codegen to access the database.
type CodegenStorage struct {
	queries *database.Queries
	db      *sql.DB
}

func NewCodegenStorage(db *sql.DB) *CodegenStorage {
	return &CodegenStorage{
		queries: database.New(db),
		db:      db,
	}
}

var _ SQLStorage = (*CodegenStorage)(nil)

func (s *CodegenStorage) Begin(ctx context.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, nil)
}

func (s *CodegenStorage) AddClient(ctx context.Context, tx *sql.Tx, client identity.Client) error {
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

func (s *CodegenStorage) GetClientByID(ctx context.Context, tx *sql.Tx, id string) (*identity.Client, error) {
	q := s.queries.WithTx(tx)
	result, err := q.GetClientByID(ctx, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &identity.Client{
		ID:     result.ClientID,
		Secret: result.ClientSecret,
	}, nil
}

func (s *CodegenStorage) DeleteClientByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteClientByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *CodegenStorage) AddAccessToken(ctx context.Context, tx *sql.Tx, refreshTokenID string, token identity.AccessToken) error {
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

func (s *CodegenStorage) GetAccessTokenByJWT(ctx context.Context, tx *sql.Tx, jwt identity.JWT) (*identity.AccessToken, error) {
	q := s.queries.WithTx(tx)
	result, err := q.GetAccessTokenByJWT(ctx, string(jwt))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &identity.AccessToken{
		ID:               result.ClientID,
		Token:            identity.JWT(result.Jwt),
		ExpiresInSeconds: result.ExpiresInSeconds,
		UserID:           result.UserID,
		Scope:            identity.Scope(result.Scope),
		IssuedAt:         result.IssuedAt,
		TokenType:        identity.TokenType(result.Type),
		ClientID:         result.ClientID,
		Revoked:          result.Revoked,
	}, nil
}

func (s *CodegenStorage) RevokeAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	_, err := q.RevokeAccessTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *CodegenStorage) DeleteAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteAccessTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *CodegenStorage) AddRefreshToken(ctx context.Context, tx *sql.Tx, token identity.RefreshToken) error {
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

func (s *CodegenStorage) GetRefreshTokenByJWT(ctx context.Context, tx *sql.Tx, refreshToken identity.JWT) (*identity.RefreshToken, error) {
	q := s.queries.WithTx(tx)
	result, err := q.GetRefreshTokenByJWT(ctx, string(refreshToken))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &identity.RefreshToken{
		ID:       result.ClientID,
		Token:    identity.JWT(result.Jwt),
		ClientID: result.ClientID,
		Revoked:  result.Revoked,
	}, nil
}

func (s *CodegenStorage) RevokeRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	_, err := q.RevokeRefreshTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *CodegenStorage) DeleteRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteRefreshTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *CodegenStorage) AddUserInfo(ctx context.Context, tx *sql.Tx, user identity.UserInfo) error {
	q := s.queries.WithTx(tx)
	_, err := q.AddUser(ctx, database.AddUserParams{
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

func (s *CodegenStorage) GetUserInfoByUsername(ctx context.Context, tx *sql.Tx, username string) (*identity.UserInfo, error) {
	q := s.queries.WithTx(tx)
	result, err := q.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &identity.UserInfo{
		ID:           result.UserID,
		Email:        result.Email,
		Username:     result.Username,
		PasswordHash: result.PasswordHash,
	}, nil
}

func (s *CodegenStorage) DeleteUserInfoByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteUserByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
