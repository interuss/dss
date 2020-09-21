package cockroach

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/interuss/stacktrace"
)

var (
	UnknownVersion = &semver.Version{}
)

type (
	// Credentials models connect credentials.
	Credentials struct {
		Username string
		Password string
	}

	// SSL models SSL configuration parameters.
	SSL struct {
		Mode string
		Dir  string
	}

	// ConnectParameters bundles up parameters used for connecting to a CRDB instance.
	ConnectParameters struct {
		ApplicationName    string
		Host               string
		Port               int
		DBName             string
		Credentials        Credentials
		SSL                SSL
		MaxOpenConns       int
		MaxIdleConns       int
		MaxConnLifeMinutes int
	}
)

func parsePortOrDefault(port string, defaultPort int64) int64 {
	p, err := strconv.ParseInt(port, 10, 16)
	if err != nil {
		p = defaultPort
	}
	return p
}

// connectParametersFromMap constructs a ConnectParameters instance from m.
func connectParametersFromMap(m map[string]string) ConnectParameters {
	return ConnectParameters{
		ApplicationName:    m["application_name"],
		DBName:             m["db_name"],
		Host:               m["host"],
		Port:               int(parsePortOrDefault(m["port"], 0)),
		MaxOpenConns:       int(parsePortOrDefault(m["max_open_conns"], 0)),
		MaxIdleConns:       int(parsePortOrDefault(m["max_idle_conns"], 2)),
		MaxConnLifeMinutes: int(parsePortOrDefault(m["max_conn_life_minutes"], 15)),
		Credentials: Credentials{
			Username: m["user"],
		},
		SSL: SSL{
			Mode: m["ssl_mode"],
			Dir:  m["ssl_dir"],
		},
	}
}

// BuildURI returns a URI built from p.
func (p ConnectParameters) BuildURI() (string, error) {
	an := p.ApplicationName
	if an == "" {
		an = "dss"
	}
	h := p.Host
	if h == "" {
		return "", stacktrace.NewError("Missing crdb hostname")
	}
	port := p.Port
	if port == 0 {
		return "", stacktrace.NewError("Missing crdb port")
	}
	u := p.Credentials.Username
	if u == "" {
		return "", stacktrace.NewError("Missing crdb user")
	}
	ssl := p.SSL.Mode
	if ssl == "" {
		return "", stacktrace.NewError("Missing crdb ssl_mode")
	}
	db := p.DBName
	if db != "" {
		db = fmt.Sprintf("/%s", db)
	}
	if ssl == "disable" {
		return fmt.Sprintf("postgresql://%s@%s:%d%s?application_name=%s&sslmode=disable", u, h, port, db, an), nil
	}
	dir := p.SSL.Dir
	if dir == "" {
		return "", stacktrace.NewError("Missing crdb ssl_dir")
	}

	return fmt.Sprintf(
		"postgresql://%s@%s:%d%s?application_name=%s&sslmode=%s&sslrootcert=%s/ca.crt&sslcert=%s/client.%s.crt&sslkey=%s/client.%s.key",
		u, h, port, db, an, ssl, dir, dir, u, dir, u,
	), nil
}

// DB models a connection to a CRDB instance.
type DB struct {
	*sql.DB
}

// Dial returns a DB instance connected to a cockroach instance available at
// "uri".
// https://www.cockroachlabs.com/docs/stable/connection-parameters.html
func Dial(connParams ConnectParameters) (*DB, error) {
	uri, err := connParams.BuildURI()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error building URI")
	}
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(connParams.MaxOpenConns)
	db.SetMaxIdleConns(connParams.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(connParams.MaxConnLifeMinutes) * time.Minute)

	return &DB{
		DB: db,
	}, nil
}

// GetVersion returns the Schema Version of the requested DB Name
func (db *DB) GetVersion(ctx context.Context, dbName string) (*semver.Version, error) {
	const query = `
		SELECT EXISTS (
			SELECT
				*
			FROM
				information_schema.tables
			WHERE
				table_name = 'schema_versions'
			AND
				table_catalog = $1
		)
	`

	var (
		exists          bool
		getVersionQuery = fmt.Sprintf(`
		SELECT
			schema_version
		FROM
			%s.schema_versions
		WHERE
			onerow_enforcer = TRUE`, dbName)
	)

	if err := db.QueryRowContext(ctx, query, dbName).Scan(&exists); err != nil {
		return nil, stacktrace.Propagate(err, "Error scanning table listing row")
	}

	if !exists {
		// Database has not been bootstrapped using DB Schema Manager
		return UnknownVersion, nil
	}

	var dbVersion string
	if err := db.QueryRowContext(ctx, getVersionQuery).Scan(&dbVersion); err != nil {
		return nil, stacktrace.Propagate(err, "Error scanning version row")
	}
	if len(dbVersion) > 0 && dbVersion[0] == 'v' {
		dbVersion = dbVersion[1:]
	}

	return semver.NewVersion(dbVersion)
}
