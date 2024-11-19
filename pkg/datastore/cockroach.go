package datastore

import (
	"context"
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5/pgxpool"
	"regexp"
)

// DB models a connection to a CRDB instance.
type DB struct {
	Pool *pgxpool.Pool
}

// GetVersion returns the Schema Version of the requested DB Name
func (db *DB) GetVersion(ctx context.Context, dbName string) (*semver.Version, error) {
	if dbName == "" {
		return nil, stacktrace.NewError("GetVersion was provided with an empty database name")
	}
	var (
		checkTableQuery = fmt.Sprintf(`
      SELECT EXISTS (
        SELECT
          *
        FROM
          %s.information_schema.tables
        WHERE
          table_name = 'schema_versions'
        AND
          table_catalog = $1
      )`, dbName)
		exists          bool
		getVersionQuery = fmt.Sprintf(`
      SELECT
        schema_version
      FROM
        %s.schema_versions
      WHERE
        onerow_enforcer = TRUE`, dbName)
	)

	if err := db.Pool.QueryRow(ctx, checkTableQuery, dbName).Scan(&exists); err != nil {
		return nil, stacktrace.Propagate(err, "Error scanning table listing row")
	}

	if !exists {
		// Database has not been bootstrapped using DB Schema Manager
		return UnknownVersion, nil
	}

	var dbVersion string
	if err := db.Pool.QueryRow(ctx, getVersionQuery).Scan(&dbVersion); err != nil {
		return nil, stacktrace.Propagate(err, "Error scanning version row")
	}
	if len(dbVersion) > 0 && dbVersion[0] == 'v' {
		dbVersion = dbVersion[1:]
	}

	return semver.NewVersion(dbVersion)
}

func (db *DB) GetServerVersion() (*semver.Version, error) {
	const versionDbQuery = `
      SELECT version();
    `
	var fullVersion string
	err := db.Pool.QueryRow(context.Background(), versionDbQuery).Scan(&fullVersion)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error querying CRDB server version")
	}

	re := regexp.MustCompile(`v((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*))`)
	match := re.FindStringSubmatch(fullVersion)
	version, err := semver.NewVersion(match[1])
	if err != nil {
		return nil, stacktrace.Propagate(err, "CRDB server version could not be parsed in semver format")
	}
	return version, nil
}
