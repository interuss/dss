package cockroach

import (
	"context"
	"errors"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/logging"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

var (
	storeURI  = flag.String("store-uri", "", "URI pointing to a Cockroach node")
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().Add(-time.Minute)
	endTime   = fakeClock.Now().Add(time.Hour)
	writer    = "writer"
)

func init() {
	DefaultTimeout = 50 * time.Millisecond
}

func setUpStore(ctx context.Context, t *testing.T) (*Store, func()) {
	if len(*storeURI) == 0 {
		t.Skip()
	} else {
		if !(strings.Contains(*storeURI, "rid") || strings.Contains(*storeURI, "scd")) {
			*storeURI = strings.Replace(*storeURI, "?sslmode", "/rid?sslmode", 1)
		}
	}
	// Reset the clock for every test.
	fakeClock = clockwork.NewFakeClock()

	store, err := newStore(ctx)
	require.NoError(t, err)
	return store, func() {
		require.NoError(t, CleanUp(ctx, store))
		require.NoError(t, store.Close())
	}
}

func newStore(ctx context.Context) (*Store, error) {
	config, err := pgxpool.ParseConfig(*storeURI)
	if err != nil {
		return nil, err
	}
	db, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return &Store{
		db:     &cockroach.DB{
			dbHandler: db,
		},
		logger: logging.Logger,
		clock:  fakeClock,
	}, nil
}

// CleanUp drops all required tables from the store, useful for testing.
func CleanUp(ctx context.Context, s *Store) error {
	const query = `
	DELETE FROM subscriptions WHERE id IS NOT NULL;
	DELETE FROM identification_service_areas WHERE id IS NOT NULL;`

	_, err := s.db.Exec(ctx, query)
	return err
}

func TestDatabaseEnsuresBeginsBeforeExpires(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	var (
		begins  = time.Now()
		expires = begins.Add(-5 * time.Minute)
	)
	_, err = repo.InsertSubscription(ctx, &ridmodels.Subscription{
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

	err := store.Transact(ctx, func(repo repos.Repository) error {
		// can query within this
		isa, err := repo.InsertISA(ctx, serviceArea)
		require.NotNil(t, isa)
		return err
	})
	require.NoError(t, err)
	// can query afterwads
	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	isa, err := repo.GetISA(ctx, serviceArea.ID)
	require.NoError(t, err)
	require.NotNil(t, isa)

	// Test the retry happens
	// 20ms, let's see how many retries we get.
	// Using a context ensures we bail out.
	ctx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	count := 0
	err = store.Transact(ctx, func(repo repos.Repository) error {
		// can query within this
		count++
		// Postgre retryable error
		return &pgconn.PgError{Code: "40001"}
	})
	require.Error(t, err)
	// Ensure it was retried.
	require.Greater(t, count, 1)
}

func TestTransactor(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	defer tearDownStore()

	subscription1 := subscriptionsPool[0].input
	subscription2 := subscriptionsPool[1].input

	txnCount := 0
	err := store.Transact(ctx, func(s1 repos.Repository) error {
		// We should get to this retry, then return nothing.
		if txnCount > 0 {
			return errors.New("already failed")
		}
		txnCount++
		err := store.Transact(ctx, func(s2 repos.Repository) error {
			subs, err := s1.SearchSubscriptions(ctx, subscription1.Cells)
			require.NoError(t, err)
			require.Len(t, subs, 0)
			subs, err = s2.SearchSubscriptions(ctx, subscription1.Cells)
			require.Len(t, subs, 0)
			require.NoError(t, err)

			// Tx1 conflicts first
			_, err = s1.InsertSubscription(ctx, subscription1)
			require.NoError(t, err)

			// Tx1 is rolled back, so tx2 can proceed.
			_, err = s2.InsertSubscription(ctx, subscription2)
			require.NoError(t, err)

			return nil
		})
		return err
	})
	require.Error(t, err)

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	subs, err := repo.SearchSubscriptions(ctx, subscription1.Cells)
	require.NoError(t, err)

	require.Len(t, subs, 1)

	s, err := repo.GetSubscription(ctx, subscription1.ID)
	require.NoError(t, err)
	require.Nil(t, s)

	s, err = repo.GetSubscription(ctx, subscription2.ID)
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

	tx1, err := store.db.Begin(ctx)
	require.NoError(t, err)
	s1 := &repo{
		ISA: &isaRepo{
			Queryable: tx1,
			logger:    logging.Logger,
		},
		Subscription: &subscriptionRepo{
			Queryable: tx1,
			logger:    logging.Logger,
			clock:     DefaultClock,
		},
	}

	tx2, err := store.db.Begin(ctx)
	require.NoError(t, err)
	s2 := &repo{
		ISA: &isaRepo{
			Queryable: tx2,
			logger:    logging.Logger,
		},
		Subscription: &subscriptionRepo{
			Queryable: tx2,
			logger:    logging.Logger,
			clock:     DefaultClock,
		},
	}

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

	require.Error(t, tx1.Commit(ctx))
	require.NoError(t, tx2.Commit(ctx))

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	subs, err = repo.SearchSubscriptions(ctx, subscription2.Cells)
	require.NoError(t, err)

	require.Len(t, subs, 1)
}
