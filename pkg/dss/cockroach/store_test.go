package cockroach

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/proto"
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
		{
			name: "a subscription with a version string",
			input: &dspb.Subscription{
				Id:    uuid.NewV4().String(),
				Owner: "me-myself-and-i",
				Callbacks: &dspb.SubscriptionCallbacks{
					IdentificationServiceAreaUrl: "https://no/place/like/home",
				},
				NotificationIndex: 42,
				Version:           "12t7ftmlhgo00",
			},
		},
	}

	serviceAreasPool = []struct {
		name  string
		input *dspb.IdentificationServiceArea
	}{
		{
			name: "a subscription without begins and expires",
			input: &dspb.IdentificationServiceArea{
				Id:         uuid.NewV4().String(),
				Owner:      "me-myself-and-i",
				FlightsUrl: "https://no/place/like/home/for/flights",
				Extents: &dspb.Volume4D{
					TimeStart: begins,
					TimeEnd:   expires,
				},
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

	_, err = store.insertSubscription(ctx, &dspb.Subscription{
		Id:    uuid.NewV4().String(),
		Owner: "me-myself-and-i",
		Callbacks: &dspb.SubscriptionCallbacks{
			IdentificationServiceAreaUrl: "https://no/place/like/home",
		},
		NotificationIndex: 42,
		Begins:            tsb,
		Expires:           tse,
	}, s2.CellUnion{})
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
			sub1, err := store.insertSubscription(ctx, r.input, s2.CellUnion{})
			require.NoError(t, err)
			require.NotNil(t, sub1)

			sub2, err := store.GetSubscription(ctx, sub1.Id)
			require.NoError(t, err)
			require.NotNil(t, sub2)

			require.Equal(t, *sub1, *sub2)
		})
	}
}

func TestStoreInsertSubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			sub1, err := store.insertSubscription(ctx, r.input, s2.CellUnion{})
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Test changes without the version differing.
			r2 := sub1
			r2.Owner = "new test owner"
			sub2, err := store.insertSubscription(ctx, r2, s2.CellUnion{})
			require.NoError(t, err)
			require.NotNil(t, sub2)
			require.Equal(t, sub2.Owner, "new test owner")

			r3 := proto.Clone(r.input).(*dspb.Subscription)
			r3.Owner = "new test owner 2"
			r3.Version = "version_mismatch"
			sub3, err := store.insertSubscription(ctx, r3, s2.CellUnion{})
			require.Error(t, err)
			require.Nil(t, sub3)

			sub4, err := store.GetSubscription(ctx, sub1.Id)
			require.NoError(t, err)
			require.NotNil(t, sub4)

			require.Equal(t, *sub2, *sub4)
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
			sub1, err := store.insertSubscription(ctx, r.input, s2.CellUnion{})
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Ensure mismatched versions return an error
			sub2, err := store.DeleteSubscription(ctx, sub1.Id, "a3cg3tcuhk000")
			require.Error(t, err)
			require.Nil(t, sub2)

			sub3, err := store.DeleteSubscription(ctx, sub1.Id, sub1.Version)
			require.NoError(t, err)
			require.NotNil(t, sub3)

			require.Equal(t, *sub1, *sub3)
		})
	}
}

func TestStoreSearchSubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	var (
		inserted = []*dspb.Subscription{}
		cells    = s2.CellUnion{
			s2.CellID(42),
			s2.CellID(84),
			s2.CellID(126),
			s2.CellID(168),
		}
		owners = []string{
			"me",
			"my",
			"self",
			"and",
			"i",
		}
	)

	for i, r := range subscriptionsPool {
		subscription := *r.input
		subscription.Owner = owners[i]

		sub1, err := store.insertSubscription(ctx, &subscription, cells[:i])
		require.NoError(t, err)
		require.NotNil(t, sub1)

		inserted = append(inserted, sub1)
	}

	for _, owner := range owners {
		found, err := store.SearchSubscriptions(ctx, cells, owner)
		require.NoError(t, err)
		require.NotNil(t, found)
		// We insert one subscription per owner. Hence, no matter how many cells are touched by the subscription,
		// the result should always be 1.
		require.Len(t, found, 1)
	}
}

func TestStoreDeleteIdentificationServiceAreas(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	var (
		insertedServiceAreas  = []*dspb.IdentificationServiceArea{}
		insertedSubscriptions = []*dspb.Subscription{}
		cells                 = s2.CellUnion{
			s2.CellID(42),
		}
	)

	for _, r := range subscriptionsPool {
		s1, err := store.insertSubscription(ctx, r.input, cells)
		require.NoError(t, err)
		require.NotNil(t, s1)

		insertedSubscriptions = append(insertedSubscriptions, s1)
	}

	for _, r := range serviceAreasPool {
		saOut, err := store.insertIdentificationServiceAreaUnchecked(ctx, r.input, cells)
		require.NoError(t, err)
		require.NotNil(t, saOut)

		insertedServiceAreas = append(insertedServiceAreas, saOut)
	}

	for _, sa := range insertedServiceAreas {
		serviceAreaOut, subscriptionsOut, err := store.DeleteIdentificationServiceArea(ctx, sa.GetId(), sa.GetOwner())
		require.NoError(t, err)
		require.NotNil(t, serviceAreaOut)
		require.NotNil(t, subscriptionsOut)
	}
}
