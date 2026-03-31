package datastore

import (
	"context"

	"github.com/interuss/dss/pkg/datastore/flags"
	"github.com/interuss/dss/pkg/datastoreutils"
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
	datastore.BaseStore[repos.Repository]
}

func NewStore(ctx context.Context, db *datastore.Datastore, logger *zap.Logger, globalLock bool) (*Store, error) {

	s := &Store{}

	base, err := datastore.NewBaseStore(ctx, db, flags.ConnectParameters().MaxRetries, func(q dssql.Queryable) repos.Repository {
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
	s.BaseStore = base
	return s, s.CheckMajorSchemaVersion(ctx, currentCrdbMajorSchemaVersion, currentYugabyteMajorSchemaVersion, db.Pool.Config().ConnConfig.Database)
}

func Dial(ctx context.Context, logger *zap.Logger, globalLock bool) (*Store, error) {

	store, err := datastoreutils.DialStore(ctx, "scd", func(db *datastore.Datastore) (*Store, error) {
		return NewStore(ctx, db, logger, globalLock)
	})

	return store, err
}
