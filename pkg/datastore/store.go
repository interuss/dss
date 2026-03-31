package datastore

import (
	"context"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/coreos/go-semver/semver"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
	"github.com/jonboulle/clockwork"
)

type BaseStore[R any] struct {
	DB           *Datastore
	Clock        clockwork.Clock
	databaseName string
	version      *semver.Version
	newRepo      func(dssql.Queryable) R
	maxRetries   int
}

func NewBaseStore[R any](ctx context.Context, db *Datastore, maxRetries int, newRepo func(dssql.Queryable) R) (BaseStore[R], error) {

	dbName := db.Pool.Config().ConnConfig.Database

	vs, err := db.GetSchemaVersion(ctx, dbName)
	if err != nil {
		return BaseStore[R]{}, stacktrace.Propagate(err, "Failed to get schema version for %s", dbName)
	}
	return BaseStore[R]{
		DB:           db,
		Clock:        clockwork.NewRealClock(),
		databaseName: dbName,
		version:      vs,
		newRepo:      newRepo,
		maxRetries:   maxRetries,
	}, nil
}

func (s *BaseStore[R]) Interact(_ context.Context) (R, error) {
	return s.newRepo(s.DB.Pool), nil
}

func (s *BaseStore[R]) Transact(ctx context.Context, f func(context.Context, R) error) error {
	ctx = crdb.WithMaxRetries(ctx, s.maxRetries)
	return crdbpgx.ExecuteTx(ctx, s.DB.Pool, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		return f(ctx, s.newRepo(tx))
	})
}

func (s *BaseStore[R]) GetVersion(ctx context.Context) (*semver.Version, error) {
	if s.version == nil {
		vs, err := s.DB.GetSchemaVersion(ctx, s.databaseName)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get schema version for %s", s.databaseName)
		}
		s.version = vs
	}
	return s.version, nil
}

func (s *BaseStore[R]) CheckMajorSchemaVersion(ctx context.Context, crdbExpected, ybExpected int64, module string) error {
	vs, err := s.GetVersion(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get schema version for %s", module)
	}
	if vs == UnknownVersion {
		return stacktrace.NewError("%s database not bootstrapped with Schema Manager", module)
	}
	if s.DB.Version.Type == CockroachDB && crdbExpected != vs.Major {
		return stacktrace.NewError("%s: unsupported schema version %s, requires major %d", module, vs, crdbExpected)
	}
	if s.DB.Version.Type == Yugabyte && ybExpected != vs.Major {
		return stacktrace.NewError("%s: unsupported schema version %s, requires major %d", module, vs, ybExpected)
	}
	return nil
}

func (s *BaseStore[R]) Close() error {
	s.DB.Pool.Close()
	return nil
}
