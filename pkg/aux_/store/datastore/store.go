package datastore

import (
	"context"

	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/datastore"
	"github.com/interuss/dss/pkg/logging"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"
)

const (
	// The current major schema version per datastore type.
	currentCrdbMajorSchemaVersion     = 1
	currentYugabyteMajorSchemaVersion = 1
)

type repo struct {
	dssql.Queryable
	clock   clockwork.Clock
	logger  *zap.Logger
	version *datastore.Version
}

// Init initializes the SQL-backed rid store. It return a concrete datastore.Store[aux_.repos.Repository] providing the
// ability to interact with a database-backed store of aux information.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool) (*datastore.Store[repos.Repository], error) {
	return datastore.Init(ctx, datastore.Config[repos.Repository]{
		DBName:                 "aux",
		CrdbMajorSchemaVersion: currentCrdbMajorSchemaVersion,
		YbMajorSchemaVersion:   currentYugabyteMajorSchemaVersion,
		NewRepo: func(q dssql.Queryable, clock clockwork.Clock, version *datastore.Version) repos.Repository {
			return &repo{
				Queryable: q,
				clock:     clock,
				logger:    logging.WithValuesFromContext(ctx, logger),
				version:   version,
			}
		},
	}, withCheckCron)
}
