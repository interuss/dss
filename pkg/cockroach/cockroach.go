package cockroach

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5/pgxpool"
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
		MaxConnIdleSeconds int
		MaxRetries         int
	}
)

// DB models a connection to a CRDB instance.
type DB struct {
	Pool *pgxpool.Pool
}

func parseIntOrDefault(port string, defaultPort int64) int64 {
	p, err := strconv.ParseInt(port, 10, 16)
	if err != nil {
		p = defaultPort
	}
	return p
}

// connectParametersFromMap constructs a ConnectParameters instance from m.
func connectParametersFromMap(m map[string]string) ConnectParameters {
	return ConnectParameters{
		ApplicationName: m["application_name"],
		DBName:          m["db_name"],
		Host:            m["host"],
		Port:            int(parseIntOrDefault(m["port"], 0)),
		Credentials: Credentials{
			Username: m["user"],
		},
		SSL: SSL{
			Mode: m["ssl_mode"],
			Dir:  m["ssl_dir"],
		},
		MaxOpenConns:       int(parseIntOrDefault(m["max_open_conns"], 4)),
		MaxConnIdleSeconds: int(parseIntOrDefault(m["max_conn_idle_secs"], 40)),
	}
}

// formatDSN constructs a DSN string from a key value map.
func formatDSN(dsnMap map[string]string) string {
	d := make([]string, 0)
	for key, value := range dsnMap {
		if value != "" {
			d = append(d, fmt.Sprintf("%s=%s", key, value))
		}
	}
	sort.Strings(d)
	return strings.Join(d, " ")
}

// BuildURI returns a URI built from p.
func (cp ConnectParameters) BuildDSN() (string, error) {
	dsnMap := make(map[string]string)

	u := cp.Credentials.Username
	if u == "" {
		return "", stacktrace.NewError("Missing crdb user")
	}
	dsnMap["user"] = u

	h := cp.Host
	if h == "" {
		return "", stacktrace.NewError("Missing crdb hostname")
	}
	dsnMap["host"] = h

	port := cp.Port
	if port == 0 {
		return "", stacktrace.NewError("Missing crdb port")
	}
	dsnMap["port"] = fmt.Sprintf("%d", port)

	an := cp.ApplicationName
	if an == "" {
		an = "dss"
	}
	dsnMap["application_name"] = an

	dsnMap["dbname"] = cp.DBName

	sslMode := cp.SSL.Mode
	if sslMode == "" {
		return "", stacktrace.NewError("Missing crdb ssl_mode")
	}
	dsnMap["sslmode"] = sslMode

	dsnMap["pool_max_conns"] = fmt.Sprintf("%d", cp.MaxOpenConns)

	if sslMode == "disable" {
		return formatDSN(dsnMap), nil
	}

	dir := cp.SSL.Dir
	if dir == "" {
		return "", stacktrace.NewError("Missing crdb ssl_dir")
	}
	dsnMap["sslrootcert"] = fmt.Sprintf("%s/ca.crt", dir)
	dsnMap["sslcert"] = fmt.Sprintf("%s/client.%s.crt", dir, u)
	dsnMap["sslkey"] = fmt.Sprintf("%s/client.%s.key", dir, u)

	return formatDSN(dsnMap), nil
}

// Dial returns a DB instance connected to a cockroach instance available at
// "uri".
// https://www.cockroachlabs.com/docs/stable/connection-parameters.html
func Dial(ctx context.Context, connParams ConnectParameters) (*DB, error) {
	dsn, err := connParams.BuildDSN()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create connection config for pgx")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to parse connection config for pgx")
	}

	if connParams.SSL.Mode == "enable" {
		config.ConnConfig.TLSConfig.ServerName = connParams.Host
	}
	config.MaxConns = int32(connParams.MaxOpenConns)
	config.MaxConnIdleTime = (time.Duration(connParams.MaxConnIdleSeconds) * time.Second)
	config.HealthCheckPeriod = (1 * time.Second)
	config.MinConns = 1

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return &DB{
		Pool: db,
	}, nil
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
