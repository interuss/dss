package application

import (
	"context"
	"testing"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/dss/pkg/rid/store"
	ridc "github.com/interuss/dss/pkg/rid/store/sqlstore"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/dss/pkg/sqlstore"
	"github.com/interuss/dss/pkg/sqlstore/params"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

var (
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().Add(-time.Minute)
	endTime   = fakeClock.Now().Add(time.Hour)
)

type mockRepo struct {
	*isaStore
	*subscriptionStore
	dssql.Queryable
}

func (s *mockRepo) Interact(ctx context.Context) (repos.Repository, error) {
	return s, nil
}

func (s *mockRepo) Transact(ctx context.Context, f func(ctx context.Context, repo repos.Repository) error) error {
	return f(ctx, s)
}

func (s *mockRepo) Close() error {
	return nil
}

func setUpStore(ctx context.Context, t *testing.T, logger *zap.Logger) (store.Store, func()) {
	DefaultClock = fakeClock
	connectParameters := params.GetConnectParameters()

	if connectParameters.Host == "" || connectParameters.Port == 0 {
		logger.Info("using the stubbed in memory store.")
		return &mockRepo{
			isaStore: &isaStore{
				isas: make(map[dssmodels.ID]*ridmodels.IdentificationServiceArea),
			},
			subscriptionStore: &subscriptionStore{
				subs: make(map[dssmodels.ID]*ridmodels.Subscription),
			},
		}, func() {}
	}

	connectParameters.DBName = "rid"

	store, err := ridc.Init(ctx, logger, false)
	require.NoError(t, err)
	logger.Info("using sqlstore.")

	store.Clock = fakeClock

	return store, func() {
		require.NoError(t, cleanUp(ctx, store))
		require.NoError(t, store.Close())
	}
}

// cleanUp drops all required tables from the store, useful for testing.
func cleanUp(ctx context.Context, s *sqlstore.Store[repos.Repository]) error {
	const query = `
    DELETE FROM subscriptions WHERE id IS NOT NULL;
    DELETE FROM identification_service_areas WHERE id IS NOT NULL;`

	_, err := s.Pool.Exec(ctx, query)
	return err

}
