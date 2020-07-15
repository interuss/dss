package cockroach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// DB models a connection to a CRDB instance.
type DB struct {
	*sql.DB
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
	db := params["db_name"]
	if db != "" {
		db = fmt.Sprintf("/%s", db)
	}
	if ssl == "disable" {
		return fmt.Sprintf("postgresql://%s@%s:%s%s?application_name=%s&sslmode=disable", u, h, p, db, an), nil
	}
	dir := params["ssl_dir"]
	if dir == "" {
		return "", errors.New("missing crdb ssl_dir")
	}

	return fmt.Sprintf(
		"postgresql://%s@%s:%s%s?application_name=%s&sslmode=%s&sslrootcert=%s/ca.crt&sslcert=%s/client.%s.crt&sslkey=%s/client.%s.key",
		u, h, p, db, an, ssl, dir, dir, u, dir, u,
	), nil
}

// GetVersion returns the Schema Version of the requested DB Name
func GetVersion(ctx context.Context, db *DB, dbName string) (string, error) {
	const query = `
		SELECT EXISTS (
			SELECT *
				FROM information_schema.tables 
			WHERE table_name = 'schema_versions'
			AND table_catalog = $1
		)
	`
	row := db.QueryRowContext(ctx, query, dbName)
	var ret bool
	if err := row.Scan(&ret); err != nil {
		return "", err
	}
	if !ret {
		// Database has not been bootstrapped using DB Schema Manager
		return "v0.0.0", fmt.Errorf("%s has not been bootstrapped with Schema Manager, Please check https://github.com/interuss/dss/tree/master/build#updgrading-database-schemas", dbName)
	}
	getVersionQuery := fmt.Sprintf(`
		SELECT schema_version 
			FROM %s.schema_versions
		WHERE onerow_enforcer = TRUE`, dbName)
	row = db.QueryRowContext(ctx, getVersionQuery)
	var dbVersion string
	if err := row.Scan(&dbVersion); err != nil {
		return "", err
	}
	return dbVersion, nil
}
