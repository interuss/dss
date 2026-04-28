package sqlstore

import (
	"context"

	dssql "github.com/interuss/dss/pkg/sql"

	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/dss/pkg/sqlstore"
	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"
)

const (
	// The current major schema version per sqlstore type.
	currentCrdbMajorSchemaVersion     = 3
	currentYugabyteMajorSchemaVersion = 1
)

// scd.store.sqlstore.repo is a full implementation of scd.repos.Repository for data backings that
// use a database such as CockroachDB or YugabyteDB.
type repo struct {
	q          dssql.Queryable
	clock      clockwork.Clock
	logger     *zap.Logger
	globalLock bool
}

// Init initializes the SQL-backed sid store. It return a concrete sqlstore.Store[sid.repos.Repository] providing the
// ability to interact with a database-backed store of sid information.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool, globalLock bool) (*sqlstore.Store[repos.Repository], error) {
	return sqlstore.Init(ctx, sqlstore.Config[repos.Repository]{
		DBName:                 "scd",
		CrdbMajorSchemaVersion: currentCrdbMajorSchemaVersion,
		YbMajorSchemaVersion:   currentYugabyteMajorSchemaVersion,
		NewRepo: func(q dssql.Queryable, clock clockwork.Clock, _ *sqlstore.Version) repos.Repository {
			return &repo{
				q:          q,
				clock:      clock,
				logger:     logging.WithValuesFromContext(ctx, logger),
				globalLock: globalLock,
			}
		},
	}, withCheckCron)
}
