package cockroach

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/cockroach"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/dpjacques/clockwork"
	"github.com/stretchr/testify/require"
)

var (
	storeURI  = flag.String("store-uri", "", "URI pointing to a Cockroach node")
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().Add(-time.Minute)
	endTime   = fakeClock.Now().Add(time.Hour)
)

func setUpStore(ctx context.Context, t *testing.T) (*Store, func()) {
	if len(*storeURI) == 0 {
		t.Skip()
	}
	// Reset the clock for every test.
	fakeClock = clockwork.NewFakeClock()

	store, err := newStore()
	require.NoError(t, err)
	require.NoError(t, store.Bootstrap(ctx))
	return store, func() {
		require.NoError(t, cleanUp(ctx, store))
		require.NoError(t, store.Close())
	}
}

func newStore() (*Store, error) {
	cdb, err := cockroach.Dial(*storeURI)
	if err != nil {
		return nil, err
	}
	return &Store{
		ISA:          &ISAStore{Queryable: cdb, clock: fakeClock, logger: zap.L()},
		Subscription: &SubscriptionStore{Queryable: cdb, clock: fakeClock, logger: zap.L()},
		db:           cdb,
		Queryable:    cdb,
	}, nil
}

// cleanUp drops all required tables from the store, useful for testing.
func cleanUp(ctx context.Context, s *Store) error {
	const query = `
	DROP TABLE IF EXISTS cells_subscriptions;
	DROP TABLE IF EXISTS subscriptions;
	DROP TABLE IF EXISTS cells_identification_service_areas;
	DROP TABLE IF EXISTS identification_service_areas;`

	_, err := s.ExecContext(ctx, query)
	return err
}

func TestStoreBootstrap(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	tearDownStore()
}

func TestDatabaseEnsuresBeginsBeforeExpires(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	defer tearDownStore()

	var (
		begins  = time.Now()
		expires = begins.Add(-5 * time.Minute)
	)
	_, err := store.InsertSubscription(ctx, &ridmodels.Subscription{
		ID:                dssmodels.ID(uuid.New().String()),
		Owner:             "me-myself-and-i",
		URL:               "https://no/place/like/home",
		NotificationIndex: 42,
		StartTime:         &begins,
		EndTime:           &expires,
	})
	require.Error(t, err)
}

func TestTxnRetrier(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	defer tearDownStore()

	err := store.InTxnRetrier(ctx, func(store repos.Repository) error {
		return store.InTxnRetrier(ctx, func(store repos.Repository) error {
			return nil
		})
	})
	require.EqualError(t, err, "cannot call InTxnRetrier within an active Txn")

	err = store.InTxnRetrier(ctx, func(store repos.Repository) error {
		// can query within this
		isa, err := store.InsertISA(ctx, serviceArea)
		require.NotNil(t, isa)
		return err
	})
	require.NoError(t, err)
	// can query afterwads
	isa, err := store.GetISA(ctx, serviceArea.ID)
	require.NoError(t, err)
	require.NotNil(t, isa)

	// Test the retry happens
	// 20ms, let's see how many retries we get.
	// Using a context ensures we bail out.
	ctx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	count := 0
	err = store.InTxnRetrier(ctx, func(store repos.Repository) error {
		// can query within this
		count++
		// Postgre retryable error
		return &pq.Error{Code: "40001"}
	})
	require.Error(t, err)
	// Ensure it was retried.
	require.Greater(t, count, 1)
}

func TestGetVersion(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer tearDownStore()
	version, err := store.GetVersion(ctx)
	require.NoError(t, err)
	_, err = semver.Parse(version)
	require.NoError(t, err)
}
