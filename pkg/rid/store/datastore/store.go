package datastore

import (
	"context"

	"github.com/interuss/dss/pkg/datastore/flags"
	"github.com/interuss/dss/pkg/datastoreutils"
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

type repo struct {
	dssql.Queryable
	clock  clockwork.Clock
	logger *zap.Logger
}

type Store struct {
	datastore.Store[repos.Repository]
}

func NewStore(ctx context.Context, db *datastore.Datastore, logger *zap.Logger) (*Store, error) {

	s := &Store{}

	base, err := datastore.NewStore(ctx, db, flags.ConnectParameters().MaxRetries, currentCrdbMajorSchemaVersion, currentYugabyteMajorSchemaVersion, func(q dssql.Queryable) repos.Repository {
		return &repo{
			Queryable: q,
			clock:     s.Clock,
			logger:    logging.WithValuesFromContext(ctx, logger),
		}
	})
	if err != nil {
		return nil, err
	}
	s.Store = base
	return s, nil
}

func Dial(ctx context.Context, logger *zap.Logger, withCheckCron bool) (*Store, error) {

	store, err := datastoreutils.DialStore(ctx, "rid", withCheckCron, func(db *datastore.Datastore) (*Store, error) {
		return NewStore(ctx, db, logger)
	})

	return store, err
}
