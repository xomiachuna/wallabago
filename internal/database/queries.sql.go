// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries.sql

package database

import (
	"context"
	"time"
)

const currentTimestamp = `-- name: CurrentTimestamp :one
SELECT CURRENT_TIMESTAMP::timestamp
`

func (q *Queries) CurrentTimestamp(ctx context.Context) (time.Time, error) {
	row := q.queryRow(ctx, q.currentTimestampStmt, currentTimestamp)
	var column_1 time.Time
	err := row.Scan(&column_1)
	return column_1, err
}
