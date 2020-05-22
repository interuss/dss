package cockroach

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/rid/repos"
	dssql "github.com/interuss/dss/pkg/sql"
)

// DB models a connection to a CRDB instance.
type DB struct {
	// The Queryable interface is what most calls happen on. Without calling
	// InTxnRetrier, Queryable is set to the same field as db.
	dssql.Queryable

	db *sql.DB
}

// InTxnRetrier supplies a new repo, that will perform all of the DB accesses
// in a Txn, and will retry any Txn's that fail due to retry-able errors
// (typically contention).
// Note: Currently the Newly supplied Repo *does not* support nested calls
// to InTxnRetrier.
func (d *DB) InTxnRetrier(f func(repo repos.Repository) error) error {
	return nil
}

// TODO: remove this function when SCD transitions to InTxnRetrier
func (d *DB) Begin() (*sql.Tx, error) {
	logging.Logger.Warn("this method is deprecated, please use InTxnRetrier")
	return d.db.Begin()
}

func (d *DB) Close() error {
	return d.db.Close()
}

// tx, err := c.Begin()
// if err != nil {
// 	return nil, err
// }
// defer recoverRollbackRepanic(ctx, tx)

// Dial returns a DB instance connected to a cockroach instance available at
// "uri".
// https://www.cockroachlabs.com/docs/stable/connection-parameters.html
func Dial(uri string) (*DB, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	return &DB{
		db: db,
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
