package raftstore

import (
	"context"

	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/raftstore"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"
)

type repo struct {
	clock  clockwork.Clock
	logger *zap.Logger
}

// Init initializes the SQL-backed rid store. It return a concrete raftstore.Store[aux_.repos.Repository] providing the
// ability to interact with a database-backed store of aux information.
func Init(ctx context.Context, logger *zap.Logger, withCheckCron bool) (*raftstore.Store[repos.Repository], error) {
	return raftstore.Init(ctx, raftstore.Config[repos.Repository]{
		NewRepo: func(q dssql.Queryable, clock clockwork.Clock) repos.Repository {
			return &repo{
				clock:  clock,
				logger: logging.WithValuesFromContext(ctx, logger),
			}
		},
	}, withCheckCron)
}
