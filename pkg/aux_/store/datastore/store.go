package datastore

import (
	"context"

	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/datastore"
	"github.com/interuss/dss/pkg/datastore/flags"
	"github.com/interuss/dss/pkg/datastoreutils"
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

type Store struct {
	datastore.BaseStore[repos.Repository]
}

func NewStore(ctx context.Context, db *datastore.Datastore, logger *zap.Logger) (*Store, error) {

	s := &Store{}

	base, err := datastore.NewBaseStore(ctx, db, flags.ConnectParameters().MaxRetries, func(q dssql.Queryable) repos.Repository {
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
	s.BaseStore = base
	return s, s.CheckMajorSchemaVersion(ctx, currentCrdbMajorSchemaVersion, currentYugabyteMajorSchemaVersion, db.Pool.Config().ConnConfig.Database)
}

func (s *Store) CleanUp(ctx context.Context) error {
	const query = `DELETE FROM dss_metadata WHERE locality IS NOT NULL;`
	_, err := s.DB.Pool.Exec(ctx, query)
	return err
}

func Dial(ctx context.Context, logger *zap.Logger) (*Store, error) {

	store, err := datastoreutils.DialStore(ctx, "aux", func(db *datastore.Datastore) (*Store, error) {
		return NewStore(ctx, db, logger)
	})

	return store, err
}
