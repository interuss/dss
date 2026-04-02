package datastore

import (
	"context"

	"github.com/interuss/dss/pkg/datastore/flags"
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

type repo struct {
	q          dssql.Queryable
	clock      clockwork.Clock
	logger     *zap.Logger
	globalLock bool
}

type Store struct {
	datastore.Store[repos.Repository]
}

func NewStore(ctx context.Context, db *datastore.Datastore, logger *zap.Logger, globalLock bool) (*Store, error) {

	s := &Store{}

	base, err := datastore.NewStore(ctx, db, flags.ConnectParameters().MaxRetries, func(q dssql.Queryable) repos.Repository {
		return &repo{
			q:          q,
			clock:      s.Clock,
			logger:     logging.WithValuesFromContext(ctx, logger),
			globalLock: globalLock,
		}
	})
	if err != nil {
		return nil, err
	}
	s.Store = base
	return s, s.CheckMajorSchemaVersion(ctx, currentCrdbMajorSchemaVersion, currentYugabyteMajorSchemaVersion, db.Pool.Config().ConnConfig.Database)
}
