package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/steeling/InterUSS-Platform/pkg/dss/models"
	"github.com/stretchr/testify/require"
)

var (
	subscriptionsPool = []struct {
		name  string
		input *models.Subscription
	}{
		{
			name: "a subscription without startTime and endTime",
			input: &models.Subscription{
				ID:                models.ID(uuid.New().String()),
				Owner:             models.Owner(uuid.New().String()),
				Url:               "https://no/place/like/home",
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription with startTime and endTime",
			input: &models.Subscription{
				ID:                models.ID(uuid.New().String()),
				Owner:             models.Owner(uuid.New().String()),
				Url:               "https://no/place/like/home",
				StartTime:         &startTime,
				EndTime:           &endTime,
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription with startTime and without endTime",
			input: &models.Subscription{
				ID:                models.ID(uuid.New().String()),
				Owner:             models.Owner(uuid.New().String()),
				Url:               "https://no/place/like/home",
				StartTime:         &startTime,
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription without startTime and with endTime",
			input: &models.Subscription{
				ID:                models.ID(uuid.New().String()),
				Owner:             models.Owner(uuid.New().String()),
				Url:               "https://no/place/like/home",
				EndTime:           &endTime,
				NotificationIndex: 42,
			},
		},
	}
)

func TestDatabaseEnsuresStartTimeBeforeEndTime(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	var (
		startTime = time.Now()
		endTime   = time.Now().Add(-5 * time.Minute)
	)

	_, err := store.InsertSubscription(ctx, &models.Subscription{
		ID:                models.ID(uuid.New().String()),
		Owner:             models.Owner(uuid.New().String()),
		Url:               "https://no/place/like/home",
		NotificationIndex: 42,
		StartTime:         &startTime,
		EndTime:           &endTime,
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
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	for _, r := range subscriptionsPool {
		t.Run(r.name, func(t *testing.T) {
			sub1, err := store.InsertSubscription(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Test changes without the version differing.
			r2 := *sub1
			r2.Url = "new url"
			sub2, err := store.InsertSubscription(ctx, &r2)
			require.NoError(t, err)
			require.NotNil(t, sub2)
			require.Equal(t, "new url", sub2.Url)

			// Test it doesn't work when Version is nil.
			r3 := *sub2
			r3.Url = "new url 2"
			r3.Version = nil
			sub3, err := store.InsertSubscription(ctx, &r3)
			require.Error(t, err)
			require.Nil(t, sub3)

			// Bad version doesn't work.
			r4 := *sub2
			r4.Url = "new url 3"
			r4.Version = models.VersionFromTime(time.Now())
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
	makeSubscription := func(cellIDs []uint64) *models.Subscription {
		s := *subscriptionsPool[0].input
		s.Owner = models.Owner("bob")
		s.ID = models.ID(uuid.New().String())

		s.Cells = make(s2.CellUnion, len(cellIDs))
		for i, id := range cellIDs {
			s.Cells[i] = s2.CellID(id)
		}
		return &s
	}

	// We should be able to insert 10 subscriptions without error.
	for i := 0; i < 10; i++ {
		ret, err := store.InsertSubscription(ctx, makeSubscription([]uint64{42, 43}))
		require.NoError(t, err)
		require.NotNil(t, ret)
	}

	// Inserting the 11th subscription will fail.
	ret, err := store.InsertSubscription(ctx, makeSubscription([]uint64{42, 43}))
	require.EqualError(t, err, "rpc error: code = ResourceExhausted desc = too many existing subscriptions in this area")
	require.Nil(t, ret)

	// Inserting a subscription in a different cell will succeed.
	ret, err = store.InsertSubscription(ctx, makeSubscription([]uint64{45}))
	require.NoError(t, err)
	require.NotNil(t, ret)

	// Inserting a subscription that overlaps with 42 or 43 will fail.
	ret, err = store.InsertSubscription(ctx, makeSubscription([]uint64{7, 43}))
	require.EqualError(t, err, "rpc error: code = ResourceExhausted desc = too many existing subscriptions in this area")
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
			sub1, err := store.InsertSubscription(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Ensure mismatched versions return an error
			v, err := models.VersionFromString("a3cg3tcuhk000")
			require.NoError(t, err)
			sub2, err := store.DeleteSubscription(ctx, sub1.ID, sub1.Owner, v)
			require.Error(t, err)
			require.Nil(t, sub2)

			// Can't delete other users data.
			sub3, err := store.DeleteSubscription(ctx, sub1.ID, "wrong owner", sub1.Version)
			require.Error(t, err)
			require.Nil(t, sub3)

			sub4, err := store.DeleteSubscription(ctx, sub1.ID, sub1.Owner, sub1.Version)
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
		inserted = []*models.Subscription{}
		cells    = s2.CellUnion{
			s2.CellID(42),
			s2.CellID(84),
			s2.CellID(126),
			s2.CellID(168),
			s2.CellID(200),
			s2.CellID(overflow),
		}
		owners = []models.Owner{
			"me",
			"my",
			"self",
			"and",
		}
	)

	for i, r := range subscriptionsPool {
		subscription := *r.input
		subscription.Owner = owners[i]
		subscription.Cells = cells[:i]
		sub1, err := store.InsertSubscription(ctx, &subscription)
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
