package gc

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/cockroach"
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
	storeURI  = flag.String("store-uri", "", "URI pointing to a Cockroach node")
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().Add(-time.Minute)
	endTime   = fakeClock.Now().Add(time.Hour)
	writer    = "writer"
)

type mockRepo struct {
	*isaStore
	*subscriptionStore
	dssql.Queryable
}

type isaStore struct {
	isas map[dssmodels.ID]*ridmodels.IdentificationServiceArea
}
// Implements repos.ISA.DeleteISA
func (store *isaStore) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	isa, ok := store.isas[isa.ID]
	if !ok {
		return nil, nil
	}
	delete(store.isas, isa.ID)

	return isa, nil
}

type subscriptionStore struct {
	subs map[dssmodels.ID]*ridmodels.Subscription
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

	// Use a real store.
	cdb, err := cockroach.Dial(*storeURI)
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

func TestDeleteExpiredISAs(t *testing.T) {
	ctx := context.Background()
	l := zap.L()
	store, tearDownStore := setUpStore(ctx, t, l)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)
}