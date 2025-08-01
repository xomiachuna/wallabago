package storage

import (
	"context"
	"database/sql"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/database"
	"github.com/pkg/errors"
)

type UserStorage interface {
	AddUser(ctx context.Context, tx *sql.Tx, user core.User) error
	TransactionStarter
}

type userStorage struct {
	queries *database.Queries
	db      *sql.DB
}

func NewUserStorage(db *sql.DB) UserStorage {
	return &userStorage{
		db:      db,
		queries: database.New(db),
	}
}

var _ UserStorage = (*userStorage)(nil)

func (s *userStorage) Begin(ctx context.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, nil)
}

func (s *userStorage) AddUser(ctx context.Context, tx *sql.Tx, user core.User) error {
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
