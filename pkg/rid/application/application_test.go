package application

import (
	"context"
	"testing"
	"time"

	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/cockroach/flags"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/dss/pkg/rid/store"
	ridcrdb "github.com/interuss/dss/pkg/rid/store/cockroach"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

var (
	fakeClock     = clockwork.NewFakeClock()
	startTime     = fakeClock.Now().Add(-time.Minute)
	endTime       = fakeClock.Now().Add(time.Hour)
	connectParams = flags.ConnectParameters()
)

type mockRepo struct {
	*isaStore
	*subscriptionStore
	dssql.Queryable
}

func (s *mockRepo) Interact(ctx context.Context) (repos.Repository, error) {
	return s, nil
}

func (s *mockRepo) Transact(ctx context.Context, f func(repo repos.Repository) error) error {
	return f(s)
}

func (s *mockRepo) Close() error {
	return nil
}

func setUpStore(ctx context.Context, t *testing.T, logger *zap.Logger) (store.Store, func()) {
	DefaultClock = fakeClock

	if connectParams.Host == "" {
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

	_, err := connectParams.BuildURI()
	require.NoError(t, err)

	ridcrdb.DefaultClock = fakeClock
	logger.Info("using cockroachDB.")

	// Use a real store.
	cdb, err := cockroach.Dial(connectParams)
	require.NoError(t, err)

	store, err := ridcrdb.NewStore(ctx, cdb, logger)
	require.NoError(t, err)

	return store, func() {
		require.NoError(t, CleanUp(ctx, store))
		require.NoError(t, store.Close())
	}
}

// CleanUp drops all required tables from the store, useful for testing.
func CleanUp(ctx context.Context, s *ridcrdb.Store) error {
	return s.CleanUp(ctx)
}
