package sql

import (
	"context"
	"database/sql"
)

var DBKey = "db_interface_key"

// Queryable abstracts common operations on sql.DB and sql.Tx instances.
type DB interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// Queryable abstracts common operations on sql.DB and sql.Tx instances.
type Queryable struct {
	db DB
}

func getDB(ctx context.Context) DB {
	dbi := ctx.Value(DBKey)
	if dbi == nil {
		// Return the default db
		return db
	}

	return dbi.(DB)
}

func (q *Queryable) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return getDB(ctx).QueryContext(ctx, query, args)
}

func (q *Queryable) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return getDB(ctx).QueryRowContext(ctx, query, args)
}

func (q *Queryable) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return getDB(ctx).ExecContext(ctx, query, args)
}
