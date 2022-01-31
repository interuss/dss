package application

import (
	"context"
	"testing"
	"time"
	"log"

	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/cockroach/flags"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/dss/pkg/rid/store"
	ridc "github.com/interuss/dss/pkg/rid/store/cockroach"
	dssql "github.com/interuss/dss/pkg/sql"
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

func (s *mockRepo) Transact(ctx context.Context, f func(repo repos.Repository) error) error {
	return f(s)
}

func (s *mockRepo) Close() error {
	return nil
}

func setUpStore(ctx context.Context, t *testing.T, logger *zap.Logger) (store.Store, func()) {
	DefaultClock = fakeClock
	connectParameters := flags.ConnectParameters()

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
	// connectParameters.DBName = ridc.DatabaseName
	log.Println("My DBName", connectParameters.DBName)
	connectParameters.DBName = "rid"
	ridc.DefaultClock = fakeClock
	ridCrdb, err := cockroach.ConnectTo(ctx, connectParameters)
	require.NoError(t, err)
	logger.Info("using cockroachDB.")

	store, err := ridc.NewStore(ctx, ridCrdb, logger)
	require.NoError(t, err)

	return store, func() {
		require.NoError(t, CleanUp(ctx, store))
		require.NoError(t, store.Close())
	}
}

// CleanUp drops all required tables from the store, useful for testing.
func CleanUp(ctx context.Context, s *ridc.Store) error {
	return s.CleanUp(ctx)
}
