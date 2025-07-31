package storage

import (
	"context"
	"database/sql"

	"github.com/andriihomiak/wallabago/internal/bootstrap"
	"github.com/andriihomiak/wallabago/internal/database"
	"github.com/pkg/errors"
)

type BoostrapSQLStorage interface {
	GetBootstrapConditions(ctx context.Context, tx *sql.Tx) ([]bootstrap.Condition, error)
	MarkBootstrapConditionSatisfied(ctx context.Context, tx *sql.Tx, condition bootstrap.ConditionName) error

	TransactionStarter
}

type GeneratedBootstrapSQLStorage struct {
	queries *database.Queries
	pool    *sql.DB
}

func NewGeneratedBootstrapSQLStorage(pool *sql.DB) *GeneratedBootstrapSQLStorage {
	return &GeneratedBootstrapSQLStorage{
		queries: database.New(pool),
		pool:    pool,
	}
}

var _ BoostrapSQLStorage = (*GeneratedBootstrapSQLStorage)(nil)

func (s *GeneratedBootstrapSQLStorage) Begin(ctx context.Context) (*sql.Tx, error) {
	return s.pool.BeginTx(ctx, nil)
}

func (s *GeneratedBootstrapSQLStorage) GetBootstrapConditions(ctx context.Context, tx *sql.Tx) ([]bootstrap.Condition, error) {
	q := s.queries.WithTx(tx)
	res, err := q.GetBoostrapConditions(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	conditions := make([]bootstrap.Condition, 0, len(res))
	for _, condition := range res {
		conditions = append(conditions, bootstrap.Condition{
			Name:      bootstrap.ConditionName(condition.ConditionName),
			Satisfied: condition.Satisfied,
		})
	}
	return conditions, nil
}

func (s *GeneratedBootstrapSQLStorage) MarkBootstrapConditionSatisfied(
	ctx context.Context, tx *sql.Tx, condition bootstrap.ConditionName,
) error {
	q := s.queries.WithTx(tx)
	_, err := q.MarkBootstrapConditionSatisfied(ctx, string(condition))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
