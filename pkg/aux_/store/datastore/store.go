package datastore

import (
	"context"

	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/datastore"
	"github.com/interuss/dss/pkg/datastore/params"
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

// aux_.store.datastore.Store is a concrete store.Store[aux_.repos.Repository] providing the
// ability to interact with a database-backed store of aux information.
type Store struct {
	datastore.Store[repos.Repository]
}

func newStore(ctx context.Context, db *datastore.Store[repos.Repository], logger *zap.Logger) (*Store, error) {

	s := &Store{}

	base, err := datastore.NewStore(ctx, db, params.GetConnectParameters().MaxRetries, currentCrdbMajorSchemaVersion, currentYugabyteMajorSchemaVersion, func(q dssql.Queryable) repos.Repository {
		return &repo{
			Queryable: q,
			clock:     s.Clock,
			logger:    logging.WithValuesFromContext(ctx, logger),
			version:   db.Version,
		}
	})
	if err != nil {
		return nil, err
	}
	s.Store = base
	return s, nil
}

func Dial(ctx context.Context, logger *zap.Logger, withCheckCron bool) (*Store, error) {

	store, err := datastore.DialStore(ctx, "aux", withCheckCron, func(db *datastore.Store[repos.Repository]) (*Store, error) {
		return newStore(ctx, db, logger)
	})

	return store, err
}
