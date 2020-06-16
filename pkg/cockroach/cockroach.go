package cockroach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	dssql "github.com/interuss/dss/pkg/sql"
)

type key int

const (
	txKey key = iota
)

// DB abstracts common operations on sql.DB and sql.Tx instances.
type DB struct {
	*sql.DB
}

func (d *DB) getQueryable(ctx context.Context) dssql.Queryable {
	dbi := ctx.Value(txKey)
	if dbi == nil {
		// Return the default db
		return d.DB
	}

	return dbi.(dssql.Queryable)
}

func InTx(ctx context.Context) bool {
	return ctx.Value(txKey) != nil
}

func SetTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func (d *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.getQueryable(ctx).QueryContext(ctx, query, args...)
}

func (d *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.getQueryable(ctx).QueryRowContext(ctx, query, args...)
}

func (d *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.getQueryable(ctx).ExecContext(ctx, query, args...)
}

// Dial returns a DB instance connected to a cockroach instance available at
// "uri".
// https://www.cockroachlabs.com/docs/stable/connection-parameters.html
func Dial(uri string) (*DB, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	return &DB{
		DB: db,
	}, nil
}

// BuildURI returns a cockroachdb connection string from a params map.
func BuildURI(params map[string]string) (string, error) {
	an := params["application_name"]
	if an == "" {
		an = "dss"
	}
	h := params["host"]
	if h == "" {
		return "", errors.New("missing crdb hostname")
	}
	p := params["port"]
	if p == "" {
		return "", errors.New("missing crdb port")
	}
	u := params["user"]
	if u == "" {
		return "", errors.New("missing crdb user")
	}
	ssl := params["ssl_mode"]
	if ssl == "" {
		return "", errors.New("missing crdb ssl_mode")
	}
	if ssl == "disable" {
		return fmt.Sprintf("postgresql://%s@%s:%s?application_name=%s&sslmode=disable", u, h, p, an), nil
	}
	dir := params["ssl_dir"]
	if dir == "" {
		return "", errors.New("missing crdb ssl_dir")
	}

	return fmt.Sprintf(
		"postgresql://%s@%s:%s?application_name=%s&sslmode=%s&sslrootcert=%s/ca.crt&sslcert=%s/client.%s.crt&sslkey=%s/client.%s.key",
		u, h, p, an, ssl, dir, dir, u, dir, u,
	), nil
}
