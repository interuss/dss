package cockroach

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/steeling/InterUSS-Platform/pkg/dss"
	"github.com/steeling/InterUSS-Platform/pkg/dss/models"

	"github.com/stretchr/testify/require"
)

var (
	// Make sure that Store implements dss.Store.
	_ dss.Store = &Store{}

	storeURI  = flag.String("store-uri", "", "URI pointing to a Cockroach node")
	tempTime  = time.Now()
	startTime = models.NullTime{Time: tempTime.AddDate(0, 0, -1), Valid: true}
	endTime   = models.NullTime{Time: tempTime.AddDate(0, 0, 1), Valid: true}
)

func init() {
	flag.Parse()
}

func setUpStore(ctx context.Context, t *testing.T) (*Store, func() error) {
	store, err := newStore()
	if err != nil {
		t.Skip(err)
	}
	require.NoError(t, store.Bootstrap(ctx))
	return store, func() error {
		return store.cleanUp(ctx)
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

	return &Store{
		DB: db,
	}, nil
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
	_, err := store.InsertSubscription(ctx, &models.Subscription{
		ID:                uuid.NewV4().String(),
		Owner:             "me-myself-and-i",
		Url:               "https://no/place/like/home",
		NotificationIndex: 42,
		StartTime:         models.NullTime{begins, true},
		EndTime:           models.NullTime{expires, true},
	})
	require.Error(t, err)
}
