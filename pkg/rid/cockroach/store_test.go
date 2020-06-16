package cockroach

import (
	"context"
	"errors"
	"flag"
	"testing"
	"time"

	"github.com/dpjacques/clockwork"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/cockroach"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
)

var (
	storeURI  = flag.String("store-uri", "", "URI pointing to a Cockroach node")
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().Add(-time.Minute)
	endTime   = fakeClock.Now().Add(time.Hour)
)

func init() {
	DefaultTimeout = 50 * time.Millisecond
}

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
		require.NoError(t, CleanUp(ctx, store))
		require.NoError(t, store.Close())
	}
}

func newStore() (*Store, error) {
	cdb, err := cockroach.Dial(*storeURI)
	if err != nil {
		return nil, err
	}
	return &Store{
		ISAStore:          &ISAStore{Queryable: cdb, logger: zap.L()},
		SubscriptionStore: &SubscriptionStore{Queryable: cdb, clock: fakeClock, logger: zap.L()},
		db:                cdb,
	}, nil
}

// CleanUp drops all required tables from the store, useful for testing.
func CleanUp(ctx context.Context, s *Store) error {
	const query = `
	DROP TABLE IF EXISTS subscriptions;
	DROP TABLE IF EXISTS identification_service_areas;`

	_, err := s.db.ExecContext(ctx, query)
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

	err := store.InTxnRetrier(ctx, func(ctx context.Context) error {
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
	err = store.InTxnRetrier(ctx, func(ctx context.Context) error {
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
	require.NoError(t, err)

	// TODO: remove the below checks when we have better schema management
	require.Equal(t, "v2", semver.Major(version))

	_, err = store.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS cells_subscriptions (id STRING PRIMARY KEY);`)
	require.NoError(t, err)

	version, err = store.GetVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, "v1", semver.Major(version))

	_, err = store.db.ExecContext(ctx, `DROP TABLE cells_subscriptions;`)
	require.NoError(t, err)

}

func TestRepository(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	defer tearDownStore()
	subscription1 := subscriptionsPool[0].input
	subscription2 := subscriptionsPool[1].input

	txnCount := 0
	err := store.InTxnRetrier(ctx, func(ctx1 context.Context) error {
		// We should get to this retry, then return nothing.
		if txnCount > 0 {
			return errors.New("already failed")
		}
		txnCount++
		err := store.InTxnRetrier(ctx, func(ctx2 context.Context) error {
			subs, err := store.SearchSubscriptions(ctx1, subscription1.Cells)
			require.NoError(t, err)
			require.Len(t, subs, 0)
			subs, err = store.SearchSubscriptions(ctx2, subscription1.Cells)
			require.Len(t, subs, 0)
			require.NoError(t, err)

			// Tx1 conflicts first
			_, err = store.InsertSubscription(ctx1, subscription1)
			require.NoError(t, err)

			// Tx1 is rolled back, so tx2 can proceed.
			_, err = store.InsertSubscription(ctx2, subscription2)
			require.NoError(t, err)

			return nil
		})
		return err
	})
	require.Error(t, err)
	subs, err := store.SearchSubscriptions(ctx, subscription1.Cells)
	require.NoError(t, err)

	require.Len(t, subs, 1)

	s, err := store.GetSubscription(ctx, subscription1.ID)
	require.NoError(t, err)
	require.Nil(t, s)

	s, err = store.GetSubscription(ctx, subscription2.ID)
	require.NoError(t, err)
	require.NotNil(t, s)

}

// Test here for posterity to demonstrate transaction semantics
func TestBasicTxn(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	defer tearDownStore()
	subscription1 := subscriptionsPool[0].input
	subscription2 := subscriptionsPool[1].input

	tx1, err := store.db.Begin()
	require.NoError(t, err)

	s1 := *(store.SubscriptionStore)
	s1.Queryable = tx1

	tx2, err := store.db.Begin()
	require.NoError(t, err)

	s2 := *(store.SubscriptionStore)
	s2.Queryable = tx2

	require.NotEqual(t, store.SubscriptionStore.Queryable, s1.Queryable)
	require.NotEqual(t, store.SubscriptionStore.Queryable, s2.Queryable)

	subs, err := s1.SearchSubscriptions(ctx, subscription1.Cells)
	require.NoError(t, err)
	require.Len(t, subs, 0)
	subs, err = s2.SearchSubscriptions(ctx, subscription1.Cells)
	require.Len(t, subs, 0)
	require.NoError(t, err)

	// Tx1 conflicts first
	sub, err := s1.InsertSubscription(ctx, subscription1)
	require.NoError(t, err)
	require.NotNil(t, sub)
	// Tx1 is rolled back, so tx2 can proceed.
	_, err = s2.InsertSubscription(ctx, subscription2)
	require.NoError(t, err)

	require.Error(t, tx1.Commit())
	require.NoError(t, tx2.Commit())

	subs, err = store.SearchSubscriptions(ctx, subscription1.Cells)
	require.NoError(t, err)

	require.Len(t, subs, 1)
}
