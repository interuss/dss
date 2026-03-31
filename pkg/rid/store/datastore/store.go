package datastore

import (
	"context"

	"github.com/interuss/dss/pkg/datastore/flags"
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
	datastore.BaseStore[repos.Repository]
}

func NewStore(ctx context.Context, db *datastore.Datastore, logger *zap.Logger) (*Store, error) {

	s := &Store{}

	base, err := datastore.NewBaseStore(ctx, db, flags.ConnectParameters().MaxRetries, func(q dssql.Queryable) repos.Repository {
		return &repo{
			Queryable: q,
			clock:     s.Clock,
			logger:    logging.WithValuesFromContext(ctx, logger),
		}
	})
	if err != nil {
		return nil, err
	}
	s.BaseStore = base
	return s, s.CheckMajorSchemaVersion(ctx, currentCrdbMajorSchemaVersion, currentYugabyteMajorSchemaVersion, db.Pool.Config().ConnConfig.Database)
}

func (s *Store) CleanUp(ctx context.Context) error {
	const query = `
	DELETE FROM subscriptions WHERE id IS NOT NULL;
	DELETE FROM identification_service_areas WHERE id IS NOT NULL;`
	_, err := s.DB.Pool.Exec(ctx, query)
	return err
}
