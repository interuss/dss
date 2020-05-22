package application

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/cockroach"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridcrdb "github.com/interuss/dss/pkg/rid/cockroach"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

var (
	storeURI  = flag.String("store-uri", "", "URI pointing to a Cockroach node")
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().Add(-time.Minute)
	endTime   = fakeClock.Now().Add(time.Hour)
)

type mockRepo struct {
	*isaStore
	*subscriptionStore
	dssql.Queryable
}

func (s *mockRepo) InTxnRetrier(ctx context.Context, f func(repo repos.Repository) error) error {
	return f(s)
}

func (s *mockRepo) Close() error {
	return nil
}

func setUpRepo(ctx context.Context, t *testing.T, logger *zap.Logger) (repos.Repository, func()) {
	DefaultClock = fakeClock

	if len(*storeURI) == 0 {
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
	ridcrdb.DefaultClock = fakeClock
	logger.Info("using cockroachDB.")

	// Use a real store.
	cdb, err := cockroach.Dial(*storeURI)
	require.NoError(t, err)

	store := ridcrdb.NewStore(cdb, logger)
	require.NoError(t, store.Bootstrap(ctx))
	return store, func() {
		require.NoError(t, cleanUp(ctx, store))
		require.NoError(t, store.Close())
	}
}

// cleanUp drops all required tables from the store, useful for testing.
func cleanUp(ctx context.Context, s *ridcrdb.Store) error {
	const query = `
	DROP TABLE IF EXISTS cells_subscriptions;
	DROP TABLE IF EXISTS subscriptions;
	DROP TABLE IF EXISTS cells_identification_service_areas;
	DROP TABLE IF EXISTS identification_service_areas;`

	_, err := s.ExecContext(ctx, query)
	return err
}
