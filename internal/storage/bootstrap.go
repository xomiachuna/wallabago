package storage

import (
	"context"
	"database/sql"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/database"
	"github.com/pkg/errors"
)

type BoostrapSQLStorage interface {
	GetBootstrapConditions(ctx context.Context, tx *sql.Tx) ([]core.Condition, error)
	MarkBootstrapConditionSatisfied(ctx context.Context, tx *sql.Tx, condition core.ConditionName) error

	TransactionStarter
}

type bootstrapSQLStorage struct {
	queries *database.Queries
	pool    *sql.DB
}

func NewBootstrapSQLStorage(pool *sql.DB) BoostrapSQLStorage {
	return &bootstrapSQLStorage{
		queries: database.New(pool),
		pool:    pool,
	}
}

var _ BoostrapSQLStorage = (*bootstrapSQLStorage)(nil)

func (s *bootstrapSQLStorage) Begin(ctx context.Context) (*sql.Tx, error) {
	return s.pool.BeginTx(ctx, nil)
}

func (s *bootstrapSQLStorage) GetBootstrapConditions(ctx context.Context, tx *sql.Tx) ([]core.Condition, error) {
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

func (s *bootstrapSQLStorage) MarkBootstrapConditionSatisfied(
	ctx context.Context, tx *sql.Tx, condition core.ConditionName,
) error {
	q := s.queries.WithTx(tx)
	_, err := q.MarkBootstrapConditionSatisfied(ctx, string(condition))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
