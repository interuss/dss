package cockroach

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/palantir/stacktrace"
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
		return "", stacktrace.NewError("Missing crdb hostname")
	}
	p := params["port"]
	if p == "" {
		return "", stacktrace.NewError("Missing crdb port")
	}
	u := params["user"]
	if u == "" {
		return "", stacktrace.NewError("Missing crdb user")
	}
	ssl := params["ssl_mode"]
	if ssl == "" {
		return "", stacktrace.NewError("Missing crdb ssl_mode")
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
		return "", stacktrace.NewError("Missing crdb ssl_dir")
	}

	return fmt.Sprintf(
		"postgresql://%s@%s:%s%s?application_name=%s&sslmode=%s&sslrootcert=%s/ca.crt&sslcert=%s/client.%s.crt&sslkey=%s/client.%s.key",
		u, h, p, db, an, ssl, dir, dir, u, dir, u,
	), nil
}

// GetVersion returns the Schema Version of the requested DB Name
func GetVersion(ctx context.Context, db *DB, dbName string) (*semver.Version, error) {
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
		return nil, stacktrace.Propagate(err, "Error scanning table listing row")
	}
	if !ret {
		// Database has not been bootstrapped using DB Schema Manager
		return semver.New("0.0.0"), nil
	}
	getVersionQuery := fmt.Sprintf(`
		SELECT schema_version
			FROM %s.schema_versions
		WHERE onerow_enforcer = TRUE`, dbName)
	row = db.QueryRowContext(ctx, getVersionQuery)
	var dbVersion string
	if err := row.Scan(&dbVersion); err != nil {
		return nil, stacktrace.Propagate(err, "Error scanning version row")
	}
	return semver.New(string(dbVersion[1:])), nil
}
