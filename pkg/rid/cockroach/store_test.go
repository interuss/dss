package cockroach

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/cockroach"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
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
	// Reset the clock for every test.
	fakeClock = clockwork.NewFakeClock()

	store, err := newStore()
	if err != nil {
		t.Skip(err)
	}
	require.NoError(t, store.Bootstrap(ctx))
	return store, func() {
		require.NoError(t, cleanUp(ctx, store))
	}
}

func newStore() (*Store, error) {
	if len(*storeURI) == 0 {
		return nil, errors.New("Missing command-line parameter store-uri")
	}

	db, err := sql.Open("postgres", *storeURI)
	if err != nil {
		return nil, err
	}

	cdb := &cockroach.DB{DB: db}

	return &Store{
		ISA:          &ISAStore{DB: cdb, clock: fakeClock, logger: zap.L()},
		Subscription: &SubscriptionStore{cdb, fakeClock, zap.L()},
		DB:           cdb,
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
