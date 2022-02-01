package application

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/interuss/dss/pkg/cockroach"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/dss/pkg/rid/store"
	ridcrdb "github.com/interuss/dss/pkg/rid/store/cockroach"
	dssql "github.com/interuss/dss/pkg/sql"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jonboulle/clockwork"
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

	config, err := pgxpool.ParseConfig(*storeURI)
	require.NoError(t, err)

	db, err := pgxpool.ConnectConfig(ctx, config)
	require.NoError(t, err)

	cdb := &cockroach.DB{
		Pool: db,
	}
	store, err := ridcrdb.NewStore(ctx, cdb, "rid", logger)
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
