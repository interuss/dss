package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/stretchr/testify/require"
)

var (
	// Ensure the struct conforms to the interface
	_                 repos.Subscription = &SubscriptionStore{}
	subscriptionsPool                    = []struct {
		name  string
		input *ridmodels.Subscription
	}{
		{
			name: "a subscription with startTime and endTime",
			input: &ridmodels.Subscription{
				ID:                dssmodels.ID(uuid.New().String()),
				Owner:             dssmodels.Owner(uuid.New().String()),
				URL:               "https://no/place/like/home",
				StartTime:         &startTime,
				EndTime:           &endTime,
				NotificationIndex: 42,
				Cells: s2.CellUnion{
					s2.CellID(uint64(overflow)),
					12494535935418957824,
				},
			},
		},
		{
			name: "a subscription without startTime and with endTime",
			input: &ridmodels.Subscription{
				ID:                dssmodels.ID(uuid.New().String()),
				Owner:             dssmodels.Owner(uuid.New().String()),
				URL:               "https://no/place/like/home",
				EndTime:           &endTime,
				NotificationIndex: 42,
				Cells: s2.CellUnion{
					12494535935418957824,
				},
			},
		},
	}
)

func TestStoreGetSubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer tearDownStore()

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			sub1, err := store.InsertSubscription(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			sub2, err := store.GetSubscription(ctx, sub1.ID)
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
	defer tearDownStore()

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			sub1, err := store.InsertSubscription(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Test changes without the version differing.
			r2 := *sub1
			r2.URL = "new url"
			sub2, err := store.InsertSubscription(ctx, &r2)
			require.NoError(t, err)
			require.NotNil(t, sub2)
			require.Equal(t, "new url", sub2.URL)

			// Test it doesn't work when Version is nil.
			r3 := *sub2
			r3.URL = "new url 2"
			r3.Version = nil
			sub3, err := store.InsertSubscription(ctx, &r3)
			require.Error(t, err)
			require.Nil(t, sub3)

			// Bad version doesn't work.
			r4 := *sub2
			r4.URL = "new url 3"
			r4.Version = dssmodels.VersionFromTime(time.Now())
			sub4, err := store.InsertSubscription(ctx, &r4)
			require.Error(t, err)
			require.Nil(t, sub4)

			sub5, err := store.GetSubscription(ctx, sub1.ID)
			require.NoError(t, err)
			require.NotNil(t, sub5)

			require.Equal(t, *sub2, *sub5)
		})
	}
}

func TestStoreDeleteSubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer tearDownStore()

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			sub1, err := store.InsertSubscription(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Ensure mismatched versions return an error
			sub1BadVersion := *sub1
			sub1BadVersion.Version, err = dssmodels.VersionFromString("a3cg3tcuhk000")
			require.NoError(t, err)
			sub2, err := store.DeleteSubscription(ctx, &sub1BadVersion)
			require.Error(t, err)
			require.Nil(t, sub2)

			// Can't delete other users data.
			sub1BadOwner := *sub1
			sub1BadOwner.Owner = "wrongOwner"

			sub3, err := store.DeleteSubscription(ctx, &sub1BadOwner)
			require.Error(t, err)
			require.Nil(t, sub3)
			sub4, err := store.DeleteSubscription(ctx, sub1)
			require.NoError(t, err)
			require.NotNil(t, sub4)

			require.Equal(t, *sub1, *sub4)
		})
	}
}

func TestStoreSearchSubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	defer tearDownStore()

	var (
		// pick an L13 value that overflows.
		overflow = uint64(17106221850767130624)

		cells = s2.CellUnion{
			s2.CellID(12494535935418957824),
			s2.CellID(12494535866699481088),
			s2.CellID(12494535901059219456),
			s2.CellID(12494535866699481088),
			s2.CellID(overflow),
		}
		owners = []dssmodels.Owner{
			"me",
			"my",
		}
	)

	for i, r := range subscriptionsPool {
		subscription := *r.input
		subscription.Owner = owners[i]
		subscription.Cells = cells[:i+1]
		sub1, err := store.InsertSubscription(ctx, &subscription)
		require.NoError(t, err)
		require.NotNil(t, sub1)
	}

	for _, owner := range owners {
		found, err := store.SearchSubscriptionsByOwner(ctx, cells, owner)
		require.NoError(t, err)
		require.NotNil(t, found)
		// We insert one subscription per owner. Hence, no matter how many cells are touched by the subscription,
		// the result should always be 1.
		require.Len(t, found, 1)
	}
}

func TestStoreExpiredSubscription(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	endTime := fakeClock.Now().Add(24 * time.Hour)
	sub := &ridmodels.Subscription{
		ID:      dssmodels.ID(uuid.New().String()),
		Owner:   dssmodels.Owner("original owner"),
		Cells:   s2.CellUnion{s2.CellID(12494535866699481088)},
		EndTime: &endTime,
	}
	_, err := store.InsertSubscription(ctx, sub)
	require.NoError(t, err)

	// The subscription's endTime is 24 hours from now.
	fakeClock.Advance(23 * time.Hour)

	// We should still be able to find the subscription by searching and by ID.
	subs, err := store.SearchSubscriptionsByOwner(ctx, sub.Cells, "original owner")
	require.NoError(t, err)
	require.Len(t, subs, 1)

	ret, err := store.GetSubscription(ctx, sub.ID)
	require.NoError(t, err)
	require.NotNil(t, &ret)

	// But now the subscription has expired.
	fakeClock.Advance(2 * time.Hour)

	subs, err = store.SearchSubscriptionsByOwner(ctx, sub.Cells, "original owner")
	require.NoError(t, err)
	require.Len(t, subs, 0)

	ret, err = store.GetSubscription(ctx, sub.ID)
	require.Nil(t, ret)
	require.Error(t, err)
}

func TestStoreSubscriptionWithNoGeoData(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	endTime := fakeClock.Now().Add(24 * time.Hour)
	sub := &ridmodels.Subscription{
		ID:      dssmodels.ID(uuid.New().String()),
		Owner:   dssmodels.Owner("original owner"),
		EndTime: &endTime,
	}
	_, err := store.InsertSubscription(ctx, sub)
	require.Error(t, err)
}
