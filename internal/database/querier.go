// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package database

import (
	"context"
	"time"
)

type Querier interface {
	CurrentTimestamp(ctx context.Context) (time.Time, error)
}

var _ Querier = (*Queries)(nil)
