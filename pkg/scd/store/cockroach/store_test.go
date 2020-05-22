package cockroach

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/cockroach"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	scdstore "github.com/interuss/dss/pkg/scd/store"

	"github.com/dpjacques/clockwork"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	// Make sure that Store implements rid.Store.
	_ scdstore.Store = &Store{}

	storeURI  = flag.String("store-uri", "", "URI pointing to a Cockroach node")
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().Add(-time.Minute)
	endTime   = fakeClock.Now().Add(time.Hour)
)

func setUpStore(ctx context.Context, t *testing.T) (*Store, func() error) {
	if len(*storeURI) == 0 {
		t.Skip()
	}
	// Reset the clock for every test.
	fakeClock = clockwork.NewFakeClock()

	store, err := newStore()
	require.NoError(t, err)
	require.NoError(t, store.Bootstrap(ctx))
	return store, func() error {
		return cleanUp(ctx, store)
	}
}

func newStore() (*Store, error) {
	cdb, err := cockroach.Dial(*storeURI)
	if err != nil {
		return nil, err
	}
	return &Store{
		DB:     cdb,
		logger: zap.L(),
		clock:  fakeClock,
	}, nil
}

// cleanUp drops all required tables from the store, useful for testing.
func cleanUp(ctx context.Context, s *Store) error {
	const query = `
	DROP TABLE IF EXISTS scd_cells_operations;
	DROP TABLE IF EXISTS scd_operations;
	DROP TABLE IF EXISTS scd_cells_subscriptions;
	DROP TABLE IF EXISTS scd_subscriptions;`

	_, err := s.ExecContext(ctx, query)
	return err
}

func TestStoreBootstrap(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	require.NoError(t, tearDownStore())
}

func TestDatabaseEnsuresBeginsBeforeExpires(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	var (
		begins  = time.Now()
		expires = begins.Add(-5 * time.Minute)
	)
	_, _, err := store.UpsertSubscription(ctx, &scdmodels.Subscription{
		ID:                   scdmodels.ID(uuid.New().String()),
		Owner:                "me-myself-and-i",
		BaseURL:              "https://no/place/like/home",
		NotificationIndex:    42,
		NotifyForConstraints: true,
		StartTime:            &begins,
		EndTime:              &expires,
		Cells: s2.CellUnion{
			s2.CellID(42),
		},
	})
	require.Error(t, err)
}

func TestDatabaseEnsuresOneNotifyFlagTrue(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	var (
		begins  = time.Now()
		expires = begins.Add(5 * time.Minute)
	)
	_, _, err := store.UpsertSubscription(ctx, &scdmodels.Subscription{
		ID:                scdmodels.ID(uuid.New().String()),
		Owner:             "me-myself-and-i",
		BaseURL:           "https://no/place/like/home",
		NotificationIndex: 42,
		StartTime:         &begins,
		EndTime:           &expires,
		Cells: s2.CellUnion{
			s2.CellID(42),
		},
	})
	require.Error(t, err)
}
