package sqlstore

import (
	"context"

	dssql "github.com/interuss/dss/pkg/sql"

	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/dss/pkg/sqlstore"
	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"
)

const (
	// The current major schema version per sqlstore type.
	currentCrdbMajorSchemaVersion     = 4
	currentYugabyteMajorSchemaVersion = 1
)

// rid.store.sqlstore.repo is a full implementation of rid.repos.Repository for data backings that
// use a database such as CockroachDB or YugabyteDB.
type repo struct {
	dssql.Queryable
	clock  clockwork.Clock
	logger *zap.Logger
}

// Init initializes the SQL-backed rid store. It return a concrete sqlstore.Store[rid.repos.Repository] providing the
// ability to interact with a database-backed store of rid information.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool) (*sqlstore.Store[repos.Repository], error) {
	return sqlstore.Init(ctx, sqlstore.Config[repos.Repository]{
		DBName:                 "rid",
		CrdbMajorSchemaVersion: currentCrdbMajorSchemaVersion,
		YbMajorSchemaVersion:   currentYugabyteMajorSchemaVersion,
		NewRepo: func(q dssql.Queryable, clock clockwork.Clock, _ *sqlstore.Version) repos.Repository {
			return &repo{
				Queryable: q,
				clock:     clock,
				logger:    logging.WithValuesFromContext(ctx, logger),
			}
		},
	}, withCheckCron)
}
