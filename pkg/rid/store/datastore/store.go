package datastore

import (
	"context"

	dssql "github.com/interuss/dss/pkg/sql"

	"github.com/interuss/dss/pkg/datastore"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"
)

const (
	// The current major schema version per datastore type.
	currentCrdbMajorSchemaVersion     = 4
	currentYugabyteMajorSchemaVersion = 1
)

// rid.store.datastore.repo is a full implementation of rid.repos.Repository for data backings that
// use a database such as CockroachDB or YugabyteDB.
type repo struct {
	dssql.Queryable
	clock  clockwork.Clock
	logger *zap.Logger
}

// Init initializes the SQL-backed rid store. It return a concrete datastore.Store[rid.repos.Repository] providing the
// ability to interact with a database-backed store of rid information.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool) (*datastore.Store[repos.Repository], error) {
	return datastore.Init(ctx, datastore.Config[repos.Repository]{
		DBName:                 "rid",
		CrdbMajorSchemaVersion: currentCrdbMajorSchemaVersion,
		YbMajorSchemaVersion:   currentYugabyteMajorSchemaVersion,
		NewRepo: func(q dssql.Queryable, clock clockwork.Clock, _ *datastore.Version) repos.Repository {
			return &repo{
				Queryable: q,
				clock:     clock,
				logger:    logging.WithValuesFromContext(ctx, logger),
			}
		},
	}, withCheckCron)
}
