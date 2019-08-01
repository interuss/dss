package cockroach

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	uuid "github.com/satori/go.uuid"
	"github.com/steeling/InterUSS-Platform/pkg/dss"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"

	"github.com/stretchr/testify/require"
)

var (
	// Make sure that Store implements dss.Store.
	_ dss.Store = &Store{}

	storeURI = flag.String("store-uri", "", "URI pointing to a Cockroach node")

	begins, expires = func() (begins *timestamp.Timestamp, expires *timestamp.Timestamp) {
		const offset = 15 * time.Second

		begins, expires = ptypes.TimestampNow(), ptypes.TimestampNow()
		begins.Seconds -= int64(offset.Seconds())
		expires.Seconds += int64(offset.Seconds())

		return begins, expires
	}()

	subscriptionsPool = []struct {
		name  string
		input *dspb.Subscription
	}{
		{
			name: "a subscription without begins and expires",
			input: &dspb.Subscription{
				Id:    uuid.NewV4().String(),
				Owner: "me-myself-and-i",
				Callbacks: &dspb.SubscriptionCallbacks{
					IdentificationServiceAreaUrl: "https://no/place/like/home",
				},
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription with begins and expires",
			input: &dspb.Subscription{
				Id:    uuid.NewV4().String(),
				Owner: "me-myself-and-i",
				Callbacks: &dspb.SubscriptionCallbacks{
					IdentificationServiceAreaUrl: "https://no/place/like/home",
				},
				Begins:            begins,
				Expires:           expires,
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription with begins and without expires",
			input: &dspb.Subscription{
				Id:    uuid.NewV4().String(),
				Owner: "me-myself-and-i",
				Callbacks: &dspb.SubscriptionCallbacks{
					IdentificationServiceAreaUrl: "https://no/place/like/home",
				},
				Begins:            begins,
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription without begins and with expires",
			input: &dspb.Subscription{
				Id:    uuid.NewV4().String(),
				Owner: "me-myself-and-i",
				Callbacks: &dspb.SubscriptionCallbacks{
					IdentificationServiceAreaUrl: "https://no/place/like/home",
				},
				Expires:           expires,
				NotificationIndex: 42,
			},
		},
	}
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

	tsb, err := ptypes.TimestampProto(begins)
	require.NoError(t, err)
	tse, err := ptypes.TimestampProto(expires)
	require.NoError(t, err)

	_, err = store.insertSubscriptionUnchecked(ctx, &dspb.Subscription{
		Id:    uuid.NewV4().String(),
		Owner: "me-myself-and-i",
		Callbacks: &dspb.SubscriptionCallbacks{
			IdentificationServiceAreaUrl: "https://no/place/like/home",
		},
		NotificationIndex: 42,
		Begins:            tsb,
		Expires:           tse,
	})
	require.Error(t, err)
}

func TestStoreGetSubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			s1, err := store.insertSubscriptionUnchecked(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, s1)

			s2, err := store.GetSubscription(ctx, s1.Id)
			require.NoError(t, err)
			require.NotNil(t, s2)

			require.Equal(t, *s1, *s2)
		})
	}
}

func TestStoreDeleteSubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			s1, err := store.insertSubscriptionUnchecked(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, s1)

			s2, err := store.DeleteSubscription(ctx, s1.Id)
			require.NoError(t, err)
			require.NotNil(t, s2)

			require.Equal(t, *s1, *s2)
		})
	}
}

func TestNullTimeScan(t *testing.T) {
	nt := &nullTime{}

	require.NoError(t, nt.Scan(nil))
	require.True(t, nt.Time.IsZero())
	require.False(t, nt.Valid)

	now := time.Now()
	require.NoError(t, nt.Scan(now))
	require.False(t, nt.Time.IsZero())
	require.True(t, nt.Valid)
	require.Equal(t, now, nt.Time)

	require.Error(t, nt.Scan("scanning from anything that is not a time.Time should fail"))
}

func TestNullTimeValue(t *testing.T) {
	nt := &nullTime{}

	value, err := nt.Value()
	require.Nil(t, value)
	require.NoError(t, err)

	now := time.Now()
	require.NoError(t, nt.Scan(now))
	value, err = nt.Value()
	require.NotNil(t, value)
	require.NoError(t, err)
	ts, ok := value.(time.Time)
	require.True(t, ok)
	require.Equal(t, now, ts)
}
