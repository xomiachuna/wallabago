package storage

import (
	"context"
	"database/sql"
	neturl "net/url"
	"time"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/core/policy"
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

func (s *PostgreSQLStorage) AddEntry(ctx context.Context, tx *sql.Tx, entry core.Entry) (*core.Entry, error) {
	q := s.queries.WithTx(tx)
	result, err := q.AddEntry(ctx, database.AddEntryParams{
		Url:     entry.URL.String(),
		Title:   entry.Title,
		OwnerID: entry.OwnerID,
		Content: entry.Content,
		Sha1:    entry.SHA1,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	entry.ID = result.EntryID
	return &entry, nil
}

func (s *PostgreSQLStorage) GetEntryBySHA1(ctx context.Context, tx *sql.Tx, ownerID string, sha1 []byte) (*core.Entry, error) {
	q := s.queries.WithTx(tx)
	entry, err := q.GetEntryBySHA1(ctx, database.GetEntryBySHA1Params{
		Sha1:    sha1,
		OwnerID: ownerID,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	url, err := neturl.Parse(entry.Url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var retrievedAt *time.Time
	if entry.RetrievedAt.Valid {
		retrievedAt = &entry.RetrievedAt.Time
	}
	return &core.Entry{
		ID:          entry.EntryID,
		URL:         *url,
		Title:       entry.Title,
		OwnerID:     entry.OwnerID,
		Content:     entry.Content,
		SHA1:        entry.Sha1,
		CreatedAt:   entry.CreatedAt,
		RetrievedAt: retrievedAt,
	}, nil
}

func (s *PostgreSQLStorage) EntryExistsBySHA1(ctx context.Context, tx *sql.Tx, ownerID string, sha1 []byte) (bool, error) {
	q := s.queries.WithTx(tx)
	_, err := q.GetEntryBySHA1(ctx, database.GetEntryBySHA1Params{
		Sha1:    sha1,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *PostgreSQLStorage) GetEntryByID(ctx context.Context, tx *sql.Tx, entryID int32) (*core.Entry, error) {
	q := s.queries.WithTx(tx)

	entry, err := q.GetEntryByID(ctx, entryID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	url, err := neturl.Parse(entry.Url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var retrievedAt *time.Time
	if entry.RetrievedAt.Valid {
		retrievedAt = &entry.RetrievedAt.Time
	}
	return &core.Entry{
		ID:          entry.EntryID,
		URL:         *url,
		Title:       entry.Title,
		OwnerID:     entry.OwnerID,
		Content:     entry.Content,
		SHA1:        entry.Sha1,
		CreatedAt:   entry.CreatedAt,
		RetrievedAt: retrievedAt,
	}, nil
}

func (s *PostgreSQLStorage) GetUserMaxScope(
	ctx context.Context, tx *sql.Tx,
	userID string, subject policy.Subject, operation policy.Operation,
) (policy.Scope, error) {
	q := s.queries.WithTx(tx)
	scopeLevel, err := q.GetUserMaxScope(ctx, database.GetUserMaxScopeParams{
		UserID:       userID,
		ResourceType: string(subject),
		Operation:    string(operation),
	})
	if err != nil {
		return policy.ScopeNone, errors.WithStack(err)
	}

	// Convert scope level to scope type
	switch scopeLevel {
	case 3:
		return policy.ScopeGlobal, nil
	case 2:
		return policy.ScopeOwn, nil
	case 1:
		return policy.ScopeNone, nil
	default: // 0 or unexpected values
		return policy.ScopeNone, nil
	}
}

func (s *PostgreSQLStorage) GetEntryOwnerID(ctx context.Context, tx *sql.Tx, entryID int32) (string, error) {
	q := s.queries.WithTx(tx)
	ownerID, err := q.GetEntryOwnerID(ctx, entryID)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return ownerID, nil
}

func (s *PostgreSQLStorage) AssignUserRole(ctx context.Context, tx *sql.Tx, userID string, role policy.Role) error {
	q := s.queries.WithTx(tx)
	err := q.AssignUserRole(ctx, database.AssignUserRoleParams{
		UserID: userID,
		RoleID: string(role),
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
