package datastore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/coreos/go-semver/semver"
	"github.com/exaring/otelpgx"
	"github.com/interuss/dss/pkg/datastore/params"
	"github.com/interuss/dss/pkg/logging"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jonboulle/clockwork"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	CodeRetryable = stacktrace.ErrorCode(1)
)

var UnknownVersion = &semver.Version{}

// Store is a partial implementation of store.Store when the data backing is a database.
// It also carries the database connection (Pool) and its Version.
type Store[R any] struct {
	Pool          *pgxpool.Pool
	Version       *Version
	Clock         clockwork.Clock
	schemaVersion *semver.Version
	newRepo       func(dsssql.Queryable) R
	maxRetries    int
}

func NewStore[R any](ctx context.Context, db *Store[R], maxRetries int, crdbExpected int64, ybExpected int64, newRepo func(dsssql.Queryable) R) (Store[R], error) {

	dbName := db.Pool.Config().ConnConfig.Database

	vs, err := db.GetSchemaVersion(ctx, dbName)
	if err != nil {
		return Store[R]{}, stacktrace.Propagate(err, "Failed to get schema version for %s", dbName)
	}

	if err := checkMajorSchemaVersion(ctx, db, vs, crdbExpected, ybExpected); err != nil {
		return Store[R]{}, err
	}

	return Store[R]{
		Pool:          db.Pool,
		Version:       db.Version,
		Clock:         clockwork.NewRealClock(),
		schemaVersion: vs,
		newRepo:       newRepo,
		maxRetries:    maxRetries,
	}, nil
}

func (s *Store[R]) Interact(_ context.Context) (R, error) {
	return s.newRepo(s.Pool), nil
}

func (s *Store[R]) Transact(ctx context.Context, f func(context.Context, R) error) error {
	ctx = crdb.WithMaxRetries(ctx, s.maxRetries)
	return crdbpgx.ExecuteTx(ctx, s.Pool, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		return f(ctx, s.newRepo(tx))
	})
}

func checkMajorSchemaVersion[R any](ctx context.Context, db *Store[R], vs *semver.Version, crdbExpected int64, ybExpected int64) error {
	if vs == UnknownVersion {
		return stacktrace.NewError("%s has not been bootstrapped with Schema Manager, please check https://github.com/interuss/dss/tree/master/build#updgrading-database-schemas", db.Pool.Config().ConnConfig.Database)
	}
	if db.Version.Type == CockroachDB && crdbExpected != vs.Major {
		return stacktrace.NewError("Unsupported schema version for %s: Got %s, requires major version of %d. Please check https://github.com/interuss/dss/tree/master/build#updgrading-database-schemas", db.Pool.Config().ConnConfig.Database, vs, crdbExpected)
	}
	if db.Version.Type == Yugabyte && ybExpected != vs.Major {
		return stacktrace.NewError("Unsupported schema version for %s: Got %s, requires major version of %d. Please check https://github.com/interuss/dss/tree/master/build#updgrading-database-schemas", db.Pool.Config().ConnConfig.Database, vs, ybExpected)
	}
	return nil
}

func (s *Store[R]) Close() error {
	s.Pool.Close()
	return nil
}

func checkDatabase[R any](ctx context.Context, db *Store[R], databaseName string) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)
	statsPtr := db.Pool.Stat()
	if int(statsPtr.TotalConns()) == 0 {
		logger.Warn("Failed periodic DB Ping (TotalConns=0)", zap.String("Database", databaseName))
	} else {
		logger.Info("Successful periodic DB Ping", zap.String("Database", databaseName))
	}
}

func DialStore[R any, S any](ctx context.Context, dbName string, withCheckCron bool, newStore func(*Store[R]) (S, error)) (S, error) {

	var zero S

	cp := params.GetConnectParameters()
	cp.DBName = dbName

	db, err := Dial[R](ctx, cp)

	if err != nil {
		if strings.Contains(err.Error(), "connect: connection refused") {
			return zero, stacktrace.PropagateWithCode(err, CodeRetryable, "Failed to connect to datastore server for %s", dbName)
		}
		return zero, stacktrace.Propagate(err, "Failed to connect to %s database", dbName)
	}
	s, err := newStore(db)
	if err != nil {
		db.Pool.Close()
		if strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), fmt.Sprintf("database \"%s\" does not exist", dbName)) || strings.Contains(err.Error(), "database has not been bootstrapped with Schema Manager") {
			return zero, stacktrace.PropagateWithCode(err, CodeRetryable, "Failed to create %s store", dbName)
		}
		return zero, stacktrace.Propagate(err, "Failed to create %s store", dbName)
	}

	if withCheckCron {
		c := cron.New()
		if _, err := c.AddFunc("@every 1m", func() { checkDatabase(ctx, db, dbName) }); err != nil {
			db.Pool.Close()
			return zero, stacktrace.Propagate(err, "Failed to schedule db check for %s", dbName)
		}
		c.Start()

		go func() {
			<-ctx.Done()
			c.Stop()
		}()
	}

	return s, nil
}

func Dial[R any](ctx context.Context, connParams params.ConnectParameters) (*Store[R], error) {
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

	config.ConnConfig.Tracer = otelpgx.NewTracer()

	dbPool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	const versionDbQuery = `
      SELECT version();
    `
	var fullVersion string
	err = dbPool.QueryRow(ctx, versionDbQuery).Scan(&fullVersion)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error querying datastore version")
	}

	version, err := NewVersion(fullVersion)

	if err != nil {
		return nil, stacktrace.Propagate(err, "Error querying datastore version")
	}

	if version.Type == CockroachDB || version.Type == Yugabyte {
		return &Store[R]{Version: version, Pool: dbPool}, nil
	}

	return nil, stacktrace.NewError("%s is not implemented yet", version.Type)
}

func (s *Store[R]) CreateDatabase(ctx context.Context, dbName string) error {
	createDB := fmt.Sprintf("CREATE DATABASE %s", dbName)
	if _, err := s.Pool.Exec(ctx, createDB); err != nil {
		return stacktrace.Propagate(err, "failed to create new database %s", dbName)
	}
	return nil
}

func (s *Store[R]) DatabaseExists(ctx context.Context, dbName string) (bool, error) {
	const checkDbQuery = `
		SELECT EXISTS (
			SELECT * FROM pg_database WHERE datname = $1
		)`

	var exists bool
	if err := s.Pool.QueryRow(ctx, checkDbQuery, dbName).Scan(&exists); err != nil {
		return false, stacktrace.Propagate(err, "Error checking %s database existence", dbName)
	}

	return exists, nil
}

// GetSchemaVersion returns the Schema Version of the requested DB Name
func (s *Store[R]) GetSchemaVersion(ctx context.Context, dbName string) (*semver.Version, error) {
	if dbName == "" {
		return nil, stacktrace.NewError("GetSchemaVersion was provided with an empty database name")
	}
	if s.Version.Type == Yugabyte && dbName != s.Pool.Config().ConnConfig.Database {
		return nil, stacktrace.NewError("Yugabyte do not support switching databases with the same connection. Unable to retrieve schema version for database %s while connected to %s.", dbName, s.Pool.Config().ConnConfig.Database)
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
		getVersionQuery = `
      SELECT
        schema_version
      FROM
        schema_versions
      WHERE
        onerow_enforcer = TRUE`
	)

	if err := s.Pool.QueryRow(ctx, checkTableQuery, dbName).Scan(&exists); err != nil {
		return nil, stacktrace.Propagate(err, "Error scanning table listing row")
	}

	if !exists {
		// Database has not been bootstrapped using DB Schema Manager
		return UnknownVersion, nil
	}

	var dbVersion string
	if err := s.Pool.QueryRow(ctx, getVersionQuery).Scan(&dbVersion); err != nil {
		return nil, stacktrace.Propagate(err, "Error scanning version row")
	}
	if len(dbVersion) > 0 && dbVersion[0] == 'v' {
		dbVersion = dbVersion[1:]
	}

	return semver.NewVersion(dbVersion)
}
