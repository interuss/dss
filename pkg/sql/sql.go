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

func getDB(ctx context.Context) (DB, error) {
	dbi := ctx.Value(DBKey)
	if dbi == nil {
		// Return the default db
		return db, nil
	}

	db, ok := dbi.(DB)
	if !ok {
		return nil, dsserr.Internal("unknown type for db key in context: %+v", db)
	}
	return db, nil
}

func (q *Queryable) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	db, err := getDB(ctx)
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args)
}

func (q *Queryable) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	db, err := getDB(ctx)
	if err != nil {
		return nil, err
	}
	return db.QueryRowContext(ctx, query, args)
}

func (q *Queryable) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db, err := getDB(ctx)
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args)
}
