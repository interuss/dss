package datastore

import (
	"context"

	"github.com/interuss/dss/pkg/datastore/params"
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

func newStore(ctx context.Context, db *datastore.Datastore, logger *zap.Logger, globalLock bool) (*Store, error) {

	s := &Store{}

	base, err := datastore.NewStore(ctx, db, params.GetConnectParameters().MaxRetries, currentCrdbMajorSchemaVersion, currentYugabyteMajorSchemaVersion, func(q dssql.Queryable) repos.Repository {
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
	return s, nil
}

func Dial(ctx context.Context, logger *zap.Logger, withCheckCron bool, globalLock bool) (*Store, error) {

	store, err := datastore.DialStore(ctx, "scd", withCheckCron, func(db *datastore.Datastore) (*Store, error) {
		return newStore(ctx, db, logger, globalLock)
	})

	return store, err
}
