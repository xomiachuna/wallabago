package storage

import (
	"context"
	"database/sql"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/database"
	"github.com/pkg/errors"
)

type PostgreSQLStorage struct {
	pool    *sql.DB
	queries *database.Queries
}

func NewPostreSQLStorage(pool *sql.DB) *PostgreSQLStorage {
	return &PostgreSQLStorage{
		pool:    pool,
		queries: database.New(pool),
	}
}

func (s *PostgreSQLStorage) Begin(ctx context.Context) (*sql.Tx, error) {
	return s.pool.BeginTx(ctx, nil)
}

func (s *PostgreSQLStorage) GetBootstrapConditions(ctx context.Context, tx *sql.Tx) ([]core.Condition, error) {
	q := s.queries.WithTx(tx)
	res, err := q.GetBoostrapConditions(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	conditions := make([]core.Condition, 0, len(res))
	for _, condition := range res {
		conditions = append(conditions, core.Condition{
			Name:      core.ConditionName(condition.ConditionName),
			Satisfied: condition.Satisfied,
		})
	}
	return conditions, nil
}

func (s *PostgreSQLStorage) MarkBootstrapConditionSatisfied(
	ctx context.Context, tx *sql.Tx, condition core.ConditionName,
) error {
	q := s.queries.WithTx(tx)
	_, err := q.MarkBootstrapConditionSatisfied(ctx, string(condition))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *PostgreSQLStorage) AddClient(ctx context.Context, tx *sql.Tx, client core.Client) error {
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

func (s *PostgreSQLStorage) GetClientByID(ctx context.Context, tx *sql.Tx, id string) (*core.Client, error) {
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

func (s *PostgreSQLStorage) DeleteClientByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteClientByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *PostgreSQLStorage) AddAccessToken(ctx context.Context, tx *sql.Tx, refreshTokenID string, token core.AccessToken) error {
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

func (s *PostgreSQLStorage) GetAccessTokenByJWT(ctx context.Context, tx *sql.Tx, jwt core.JWT) (*core.AccessToken, error) {
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

func (s *PostgreSQLStorage) RevokeAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	_, err := q.RevokeAccessTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *PostgreSQLStorage) DeleteAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteAccessTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *PostgreSQLStorage) AddRefreshToken(ctx context.Context, tx *sql.Tx, token core.RefreshToken) error {
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

func (s *PostgreSQLStorage) GetRefreshTokenByJWT(ctx context.Context, tx *sql.Tx, refreshToken core.JWT) (*core.RefreshToken, error) {
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

func (s *PostgreSQLStorage) RevokeRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	_, err := q.RevokeRefreshTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *PostgreSQLStorage) DeleteRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteRefreshTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *PostgreSQLStorage) AddUserInfo(ctx context.Context, tx *sql.Tx, user core.UserInfo) error {
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

func (s *PostgreSQLStorage) GetUserInfoByUsername(ctx context.Context, tx *sql.Tx, username string) (*core.UserInfo, error) {
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

func (s *PostgreSQLStorage) DeleteUserInfoByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteIdentityUserByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *PostgreSQLStorage) AddUser(ctx context.Context, tx *sql.Tx, user core.User) error {
	q := s.queries.WithTx(tx)
	_, err := q.AddAppUser(ctx, database.AddAppUserParams{
		UserID:   user.ID,
		IsAdmin:  user.IsAdmin,
		Username: user.Username,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
