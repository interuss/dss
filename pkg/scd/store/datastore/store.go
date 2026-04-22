package datastore

import (
	"context"

	dssql "github.com/interuss/dss/pkg/sql"

	"github.com/interuss/dss/pkg/datastore"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"
)

const (
	// The current major schema version per datastore type.
	currentCrdbMajorSchemaVersion     = 3
	currentYugabyteMajorSchemaVersion = 1
)

// scd.store.datastore.repo is a full implementation of scd.repos.Repository for data backings that
// use a database such as CockroachDB or YugabyteDB.
type repo struct {
	q          dssql.Queryable
	clock      clockwork.Clock
	logger     *zap.Logger
	globalLock bool
}

// Init initializes the SQL-backed sid store. It return a concrete datastore.Store[sid.repos.Repository] providing the
// ability to interact with a database-backed store of sid information.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool, globalLock bool) (*datastore.Store[repos.Repository], error) {
	return datastore.Init(ctx, datastore.Config[repos.Repository]{
		DBName:                 "scd",
		CrdbMajorSchemaVersion: currentCrdbMajorSchemaVersion,
		YbMajorSchemaVersion:   currentYugabyteMajorSchemaVersion,
		NewRepo: func(q dssql.Queryable, clock clockwork.Clock, _ *datastore.Version) repos.Repository {
			return &repo{
				q:          q,
				clock:      clock,
				logger:     logging.WithValuesFromContext(ctx, logger),
				globalLock: globalLock,
			}
		},
	}, withCheckCron)
}
