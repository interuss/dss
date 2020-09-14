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
	_                 repos.Subscription = &subscriptionRepo{}
	subscriptionsPool                    = []struct {
		name  string
		input *ridmodels.Subscription
	}{
		{
			name: "a subscription with startTime and endTime",
			input: &ridmodels.Subscription{
				ID:                dssmodels.ID(uuid.New().String()),
				Owner:             "myself",
				URL:               "https://no/place/like/home",
				StartTime:         &startTime,
				EndTime:           &endTime,
				NotificationIndex: 42,
				Writer:            writer,
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
				Owner:             "myself",
				URL:               "https://no/place/like/home",
				EndTime:           &endTime,
				NotificationIndex: 42,
				Cells: s2.CellUnion{
					12494535935418957824,
				},
			},
		},
		{
			name: "a subscription without startTime and with endTime",
			input: &ridmodels.Subscription{
				ID:                dssmodels.ID(uuid.New().String()),
				Owner:             "me",
				URL:               "https://no/place/like/home",
				StartTime:         &startTime,
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

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			sub1, err := repo.InsertSubscription(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			sub2, err := repo.GetSubscription(ctx, sub1.ID)
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

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			sub1, err := repo.InsertSubscription(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Test changes without the version differing.
			r2 := *sub1
			r2.URL = "new url"
			sub2, err := repo.UpdateSubscription(ctx, &r2)
			require.NoError(t, err)
			require.NotNil(t, sub2)
			require.Equal(t, "new url", sub2.URL)

			// Test it doesn't work when Version is nil.
			r3 := *sub2
			r3.URL = "new url 2"
			r3.Version = nil
			sub3, err := repo.UpdateSubscription(ctx, &r3)
			require.NoError(t, err)
			require.Nil(t, sub3)

			// Bad version doesn't work.
			r4 := *sub2
			r4.URL = "new url 3"
			r4.Version = dssmodels.VersionFromTime(time.Now())
			sub4, err := repo.UpdateSubscription(ctx, &r4)
			require.NoError(t, err)
			require.Nil(t, sub4)

			sub5, err := repo.GetSubscription(ctx, sub1.ID)
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

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			sub1, err := repo.InsertSubscription(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Ensure mismatched versions returns nothing
			sub1BadVersion := *sub1
			sub1BadVersion.Version, err = dssmodels.VersionFromString("a3cg3tcuhk000")
			require.NoError(t, err)
			sub2, err := repo.DeleteSubscription(ctx, &sub1BadVersion)
			require.NoError(t, err)
			require.Nil(t, sub2)

			sub4, err := repo.DeleteSubscription(ctx, sub1)
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

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

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
			"self",
		}
	)

	for i, r := range subscriptionsPool {
		subscription := *r.input
		subscription.Owner = owners[i]
		subscription.Cells = cells[:i+1]
		sub1, err := repo.InsertSubscription(ctx, &subscription)
		require.NoError(t, err)
		require.NotNil(t, sub1)
	}
	// Test normal search
	found, err := repo.SearchSubscriptions(ctx, cells)
	require.NoError(t, err)
	require.Len(t, found, 3)
	for _, owner := range owners {
		found, err := repo.SearchSubscriptionsByOwner(ctx, cells, owner)
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

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	endTime := fakeClock.Now().Add(24 * time.Hour)
	sub := &ridmodels.Subscription{
		ID:      dssmodels.ID(uuid.New().String()),
		Owner:   dssmodels.Owner("original owner"),
		Cells:   s2.CellUnion{s2.CellID(12494535866699481088)},
		EndTime: &endTime,
	}
	_, err = repo.InsertSubscription(ctx, sub)
	require.NoError(t, err)

	// The subscription's endTime is 24 hours from now.
	fakeClock.Advance(23 * time.Hour)

	// We should still be able to find the subscription by searching and by ID.
	subs, err := repo.SearchSubscriptionsByOwner(ctx, sub.Cells, "original owner")
	require.NoError(t, err)
	require.Len(t, subs, 1)

	ret, err := repo.GetSubscription(ctx, sub.ID)
	require.NoError(t, err)
	require.NotNil(t, &ret)

	// But now the subscription has expired.
	fakeClock.Advance(2 * time.Hour)

	subs, err = repo.SearchSubscriptionsByOwner(ctx, sub.Cells, "original owner")
	require.NoError(t, err)
	require.Len(t, subs, 0)

	ret, err = repo.GetSubscription(ctx, sub.ID)
	require.NotNil(t, ret)
	require.NoError(t, err)
}

func TestStoreSubscriptionWithNoGeoData(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	endTime := fakeClock.Now().Add(24 * time.Hour)
	sub := &ridmodels.Subscription{
		ID:      dssmodels.ID(uuid.New().String()),
		Owner:   dssmodels.Owner("original owner"),
		EndTime: &endTime,
	}
	_, err = repo.InsertSubscription(ctx, sub)
	require.Error(t, err)
}

func TestMaxSubscriptionCountInCellsByOwner(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	for _, s := range subscriptionsPool {
		_, err := repo.InsertSubscription(ctx, s.input)
		require.NoError(t, err)
	}

	count, err := repo.MaxSubscriptionCountInCellsByOwner(ctx, s2.CellUnion{12494535935418957824}, "myself")
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestListExpiredSubscriptions(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	// Insert Subscription with endtime 1 day from now
	validSubscription := *subscriptionsPool[0].input
	endTime := fakeClock.Now().Add(24 * time.Hour)
	validSubscription.EndTime = &endTime
	saOut, err := repo.InsertSubscription(ctx, &validSubscription)
	require.NoError(t, err)
	require.NotNil(t, saOut)

	// Insert Subscription with endtime 30 minutes from now
	copy := *subscriptionsPool[0].input
	endTime = fakeClock.Now().Add(30 * time.Minute)
	copy.EndTime = &endTime
	copy.ID = dssmodels.ID(uuid.New().String())
	isa, err := repo.InsertSubscription(ctx, &copy)
	require.NoError(t, err)
	require.NotNil(t, isa)

	// Set Subscription's deleted time to 30 minutes from endTime.
	expiredTime := fakeClock.Now().Add(1 * time.Hour)

	subscriptions, err := repo.ListExpiredSubscriptions(ctx, serviceArea.Cells, writer, &expiredTime)
	require.NoError(t, err)
	require.Len(t, subscriptions, 1)
}
