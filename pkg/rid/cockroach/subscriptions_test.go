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
					s2.CellID(42),
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
					s2.CellID(uint64(overflow)),
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
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			sub1, err := store.Subscription.Insert(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			sub2, err := store.Subscription.Get(ctx, sub1.ID)
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
			sub1, err := store.Subscription.Insert(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Test changes without the version differing.
			r2 := *sub1
			r2.URL = "new url"
			sub2, err := store.Subscription.Insert(ctx, &r2)
			require.NoError(t, err)
			require.NotNil(t, sub2)
			require.Equal(t, "new url", sub2.URL)

			// Test it doesn't work when Version is nil.
			r3 := *sub2
			r3.URL = "new url 2"
			r3.Version = nil
			sub3, err := store.Subscription.Insert(ctx, &r3)
			require.Error(t, err)
			require.Nil(t, sub3)

			// Bad version doesn't work.
			r4 := *sub2
			r4.URL = "new url 3"
			r4.Version = dssmodels.VersionFromTime(time.Now())
			sub4, err := store.Subscription.Insert(ctx, &r4)
			require.Error(t, err)
			require.Nil(t, sub4)

			sub5, err := store.Subscription.Get(ctx, sub1.ID)
			require.NoError(t, err)
			require.NotNil(t, sub5)

			require.Equal(t, *sub2, *sub5)
		})
	}
}

func TestStoreInsertTooManySubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)

	defer func() {
		require.NoError(t, tearDownStore())
	}()

	// Helper function that makes a subscription with a random ID, fixed owner,
	// and provided cellIDs.
	makeSubscription := func(cellIDs []uint64) *ridmodels.Subscription {
		s := *subscriptionsPool[0].input
		s.Owner = dssmodels.Owner("bob")
		s.ID = dssmodels.ID(uuid.New().String())

		s.Cells = make(s2.CellUnion, len(cellIDs))
		for i, id := range cellIDs {
			s.Cells[i] = s2.CellID(id)
		}
		return &s
	}

	// We should be able to insert 10 subscriptions without error.
	for i := 0; i < 10; i++ {
		ret, err := store.Subscription.Insert(ctx, makeSubscription([]uint64{42, 43}))
		require.NoError(t, err)
		require.NotNil(t, &ret)
	}

	// Inserting the 11th subscription will fail.
	ret, err := store.Subscription.Insert(ctx, makeSubscription([]uint64{42, 43}))
	require.EqualError(t, err, "rpc error: code = ResourceExhausted desc = too many existing subscriptions in this area already")
	require.Nil(t, ret)

	// Inserting a subscription in a different cell will succeed.
	ret, err = store.Subscription.Insert(ctx, makeSubscription([]uint64{45}))
	require.NoError(t, err)
	require.NotNil(t, &ret)

	// Inserting a subscription that overlaps with 42 or 43 will fail.
	ret, err = store.Subscription.Insert(ctx, makeSubscription([]uint64{7, 43}))
	require.EqualError(t, err, "rpc error: code = ResourceExhausted desc = too many existing subscriptions in this area already")
	require.Nil(t, ret)
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
			sub1, err := store.Subscription.Insert(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Ensure mismatched versions return an error
			sub1BadVersion := *sub1
			sub1BadVersion.Version, err = dssmodels.VersionFromString("a3cg3tcuhk000")
			require.NoError(t, err)
			sub2, err := store.Subscription.Delete(ctx, &sub1BadVersion)
			require.Error(t, err)
			require.Nil(t, sub2)

			// Can't delete other users data.
			sub1BadOwner := *sub1
			sub1BadOwner.Owner = "wrongOwner"

			sub3, err := store.Subscription.Delete(ctx, &sub1BadOwner)
			require.Error(t, err)
			require.Nil(t, sub3)
			sub4, err := store.Subscription.Delete(ctx, sub1)
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
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	var (
		overflow = -1
		cells    = s2.CellUnion{
			s2.CellID(42),
			s2.CellID(84),
			s2.CellID(126),
			s2.CellID(168),
			s2.CellID(200),
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
		sub1, err := store.Subscription.Insert(ctx, &subscription)
		require.NoError(t, err)
		require.NotNil(t, sub1)
	}

	for _, owner := range owners {
		found, err := store.Subscription.SearchByOwner(ctx, cells, owner)
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
	defer func() {
		require.NoError(t, tearDownStore())
	}()
	endTime := fakeClock.Now().Add(24 * time.Hour)
	sub := &ridmodels.Subscription{
		ID:      dssmodels.ID(uuid.New().String()),
		Owner:   dssmodels.Owner("original owner"),
		Cells:   s2.CellUnion{s2.CellID(42)},
		EndTime: &endTime,
	}
	_, err := store.Subscription.Insert(ctx, sub)
	require.NoError(t, err)

	// The subscription's endTime is 24 hours from now.
	fakeClock.Advance(23 * time.Hour)

	// We should still be able to find the subscription by searching and by ID.
	subs, err := store.Subscription.SearchByOwner(ctx, sub.Cells, "original owner")
	require.NoError(t, err)
	require.Len(t, subs, 1)

	ret, err := store.Subscription.Get(ctx, sub.ID)
	require.NoError(t, err)
	require.NotNil(t, &ret)

	// But now the subscription has expired.
	fakeClock.Advance(2 * time.Hour)

	subs, err = store.Subscription.SearchByOwner(ctx, sub.Cells, "original owner")
	require.NoError(t, err)
	require.Len(t, subs, 0)

	ret, err = store.Subscription.Get(ctx, sub.ID)
	require.Nil(t, ret)
	require.Error(t, err)
}

func TestStoreSubscriptionWithNoGeoData(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer func() {
		require.NoError(t, tearDownStore())
	}()
	endTime := fakeClock.Now().Add(24 * time.Hour)
	sub := &ridmodels.Subscription{
		ID:      dssmodels.ID(uuid.New().String()),
		Owner:   dssmodels.Owner("original owner"),
		EndTime: &endTime,
	}
	_, err := store.Subscription.Insert(ctx, sub)
	require.Error(t, err)
}
