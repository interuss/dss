package sql

import (
	"context"
	// "database/sql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// Queryable abstracts common operations on sql.DB and sql.Tx instances.
type Queryable interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
}
