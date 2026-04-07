package datastore

import (
	"context"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/coreos/go-semver/semver"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
	"github.com/jonboulle/clockwork"
)

type Store[R any] struct {
	DB         *Datastore
	Clock      clockwork.Clock
	version    *semver.Version
	newRepo    func(dsssql.Queryable) R
	maxRetries int
}

func NewStore[R any](ctx context.Context, db *Datastore, maxRetries int, crdbExpected int64, ybExpected int64, newRepo func(dsssql.Queryable) R) (Store[R], error) {

	dbName := db.Pool.Config().ConnConfig.Database

	vs, err := db.GetSchemaVersion(ctx, dbName)
	if err != nil {
		return Store[R]{}, stacktrace.Propagate(err, "Failed to get schema version for %s", dbName)
	}

	err = checkMajorSchemaVersion(ctx, db, vs, crdbExpected, ybExpected)

	if err != nil {
		return Store[R]{}, err
	}

	return Store[R]{
		DB:         db,
		Clock:      clockwork.NewRealClock(),
		version:    vs,
		newRepo:    newRepo,
		maxRetries: maxRetries,
	}, nil
}

func (s *Store[R]) Interact(_ context.Context) (R, error) {
	return s.newRepo(s.DB.Pool), nil
}

func (s *Store[R]) Transact(ctx context.Context, f func(context.Context, R) error) error {
	ctx = crdb.WithMaxRetries(ctx, s.maxRetries)
	return crdbpgx.ExecuteTx(ctx, s.DB.Pool, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		return f(ctx, s.newRepo(tx))
	})
}

func checkMajorSchemaVersion(ctx context.Context, db *Datastore, vs *semver.Version, crdbExpected int64, ybExpected int64) error {
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
	s.DB.Pool.Close()
	return nil
}
