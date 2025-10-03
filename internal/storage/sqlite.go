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

type SQLiteStorage struct {
	pool    *sql.DB
	queries *database.Queries
}

func NewSQLiteStorage(pool *sql.DB) *SQLiteStorage {
	return &SQLiteStorage{
		pool:    pool,
		queries: database.New(pool),
	}
}

func (s *SQLiteStorage) Begin(ctx context.Context) (*sql.Tx, error) {
	return s.pool.BeginTx(ctx, nil)
}

func (s *SQLiteStorage) GetBootstrapConditions(ctx context.Context, tx *sql.Tx) ([]core.Condition, error) {
	q := s.queries.WithTx(tx)
	res, err := q.GetBoostrapConditions(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	conditions := make([]core.Condition, 0, len(res))
	for _, condition := range res {
		conditions = append(conditions, core.Condition{
			Name:      core.ConditionName(condition.ConditionName),
			Satisfied: condition.Satisfied != 0,
		})
	}
	return conditions, nil
}

func (s *SQLiteStorage) MarkBootstrapConditionSatisfied(
	ctx context.Context, tx *sql.Tx, condition core.ConditionName,
) error {
	q := s.queries.WithTx(tx)
	_, err := q.MarkBootstrapConditionSatisfied(ctx, string(condition))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLiteStorage) AddClient(ctx context.Context, tx *sql.Tx, client core.Client) error {
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

func (s *SQLiteStorage) GetClientByID(ctx context.Context, tx *sql.Tx, id string) (*core.Client, error) {
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

func (s *SQLiteStorage) DeleteClientByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteClientByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLiteStorage) AddAccessToken(ctx context.Context, tx *sql.Tx, refreshTokenID string, token core.AccessToken) error {
	q := s.queries.WithTx(tx)
	_, err := q.AddAccessToken(ctx, database.AddAccessTokenParams{
		TokenID:  token.ID,
		ClientID: token.ClientID,
		Jwt:      string(token.Token),
		UserID:   token.UserID,
		Revoked:  boolToInt64(token.Revoked),
		RefreshTokenID: sql.NullString{
			Valid:  refreshTokenID != "",
			String: refreshTokenID,
		},
		Type:             string(token.TokenType),
		Scope:            string(token.Scope),
		IssuedAtUnix:     token.IssuedAt.Unix(),
		ExpiresInSeconds: token.ExpiresInSeconds,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLiteStorage) GetAccessTokenByJWT(ctx context.Context, tx *sql.Tx, jwt core.JWT) (*core.AccessToken, error) {
	q := s.queries.WithTx(tx)
	result, err := q.GetAccessTokenByJWT(ctx, string(jwt))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &core.AccessToken{
		ID:               result.TokenID,
		Token:            core.JWT(result.Jwt),
		ExpiresInSeconds: result.ExpiresInSeconds,
		UserID:           result.UserID,
		Scope:            core.Scope(result.Scope),
		IssuedAt:         time.Unix(result.IssuedAtUnix, 0),
		TokenType:        core.TokenType(result.Type),
		ClientID:         result.ClientID,
		Revoked:          result.Revoked != 0,
	}, nil
}

func (s *SQLiteStorage) RevokeAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	_, err := q.RevokeAccessTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLiteStorage) DeleteAccessTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteAccessTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLiteStorage) AddRefreshToken(ctx context.Context, tx *sql.Tx, token core.RefreshToken) error {
	q := s.queries.WithTx(tx)
	_, err := q.AddRefreshToken(ctx, database.AddRefreshTokenParams{
		TokenID:  token.ID,
		Jwt:      string(token.Token),
		ClientID: token.ClientID,
		Revoked:  boolToInt64(token.Revoked),
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLiteStorage) GetRefreshTokenByJWT(ctx context.Context, tx *sql.Tx, refreshToken core.JWT) (*core.RefreshToken, error) {
	q := s.queries.WithTx(tx)
	result, err := q.GetRefreshTokenByJWT(ctx, string(refreshToken))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &core.RefreshToken{
		ID:       result.TokenID,
		Token:    core.JWT(result.Jwt),
		ClientID: result.ClientID,
		Revoked:  result.Revoked != 0,
	}, nil
}

func (s *SQLiteStorage) RevokeRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	_, err := q.RevokeRefreshTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLiteStorage) DeleteRefreshTokenByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteRefreshTokenByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLiteStorage) AddUserInfo(ctx context.Context, tx *sql.Tx, user core.UserInfo) error {
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

func (s *SQLiteStorage) GetUserInfoByUsername(ctx context.Context, tx *sql.Tx, username string) (*core.UserInfo, error) {
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

func (s *SQLiteStorage) DeleteUserInfoByID(ctx context.Context, tx *sql.Tx, id string) error {
	q := s.queries.WithTx(tx)
	err := q.DeleteIdentityUserByID(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLiteStorage) AddUser(ctx context.Context, tx *sql.Tx, user core.User) error {
	q := s.queries.WithTx(tx)
	_, err := q.AddAppUser(ctx, database.AddAppUserParams{
		UserID:   user.ID,
		IsAdmin:  boolToInt64(user.IsAdmin),
		Username: user.Username,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *SQLiteStorage) AddEntry(ctx context.Context, tx *sql.Tx, entry core.Entry) (*core.Entry, error) {
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
	entry.CreatedAt = time.Unix(result.CreatedAtUnix, 0)
	if result.RetrievedAtUnix.Valid {
		t := time.Unix(result.RetrievedAtUnix.Int64, 0)
		entry.RetrievedAt = &t
	}
	return &entry, nil
}

func (s *SQLiteStorage) GetEntryBySHA1(ctx context.Context, tx *sql.Tx, ownerID string, sha1 []byte) (*core.Entry, error) {
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
	if entry.RetrievedAtUnix.Valid {
		t := time.Unix(entry.RetrievedAtUnix.Int64, 0)
		retrievedAt = &t
	}

	return &core.Entry{
		ID:          entry.EntryID,
		URL:         *url,
		Title:       entry.Title,
		OwnerID:     entry.OwnerID,
		Content:     entry.Content,
		SHA1:        entry.Sha1,
		CreatedAt:   time.Unix(entry.CreatedAtUnix, 0),
		RetrievedAt: retrievedAt,
	}, nil
}

func (s *SQLiteStorage) EntryExistsBySHA1(ctx context.Context, tx *sql.Tx, ownerID string, sha1 []byte) (bool, error) {
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

func (s *SQLiteStorage) GetEntryByID(ctx context.Context, tx *sql.Tx, entryID int64) (*core.Entry, error) {
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
	if entry.RetrievedAtUnix.Valid {
		t := time.Unix(entry.RetrievedAtUnix.Int64, 0)
		retrievedAt = &t
	}

	return &core.Entry{
		ID:          entry.EntryID,
		URL:         *url,
		Title:       entry.Title,
		OwnerID:     entry.OwnerID,
		Content:     entry.Content,
		SHA1:        entry.Sha1,
		CreatedAt:   time.Unix(entry.CreatedAtUnix, 0),
		RetrievedAt: retrievedAt,
	}, nil
}

func (s *SQLiteStorage) GetUserMaxScope(
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
	// The result is interface{} from sqlc, need to convert to int64
	var level int64
	switch v := scopeLevel.(type) {
	case int64:
		level = v
	case int:
		level = int64(v)
	default:
		level = 0
	}

	switch level {
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

func (s *SQLiteStorage) GetEntryOwnerID(ctx context.Context, tx *sql.Tx, entryID int64) (string, error) {
	q := s.queries.WithTx(tx)
	ownerID, err := q.GetEntryOwnerID(ctx, entryID)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return ownerID, nil
}

func (s *SQLiteStorage) AssignUserRole(ctx context.Context, tx *sql.Tx, userID string, role policy.Role) error {
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

// Helper functions

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
