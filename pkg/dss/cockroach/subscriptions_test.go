package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	uuid "github.com/satori/go.uuid"
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
				ID:                uuid.NewV4().String(),
				Owner:             "me-myself-and-i",
				Url:               "https://no/place/like/home",
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription with startTime and endTime",
			input: &models.Subscription{
				ID:                uuid.NewV4().String(),
				Owner:             "me-myself-and-i",
				Url:               "https://no/place/like/home",
				StartTime:         startTime,
				EndTime:           endTime,
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription with startTime and without endTime",
			input: &models.Subscription{
				ID:                uuid.NewV4().String(),
				Owner:             "me-myself-and-i",
				Url:               "https://no/place/like/home",
				StartTime:         startTime,
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription without startTime and with endTime",
			input: &models.Subscription{
				ID:                uuid.NewV4().String(),
				Owner:             "me-myself-and-i",
				Url:               "https://no/place/like/home",
				EndTime:           endTime,
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription with a version string",
			input: &models.Subscription{
				ID:                uuid.NewV4().String(),
				Owner:             "me-myself-and-i",
				Url:               "https://no/place/like/home",
				NotificationIndex: 42,
				UpdatedAt:         startTime.Time,
			},
		},
		{
			name: "a subscription with a different owner",
			input: &models.Subscription{
				ID:                uuid.NewV4().String(),
				Owner:             "you",
				Url:               "https://no/place/like/home",
				NotificationIndex: 42,
				UpdatedAt:         startTime.Time,
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
		startTime = models.NullTime{Time: time.Now(), Valid: true}
		endTime   = models.NullTime{Time: time.Now().Add(-5 * time.Minute), Valid: true}
	)

	_, err := store.InsertSubscription(ctx, &models.Subscription{
		ID:                uuid.NewV4().String(),
		Owner:             "me-myself-and-i",
		Url:               "https://no/place/like/home",
		NotificationIndex: 42,
		StartTime:         startTime,
		EndTime:           endTime,
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
			r2.Owner = "new test owner"
			sub2, err := store.InsertSubscription(ctx, &r2)
			require.Error(t, err)
			require.Nil(t, sub2)

			sub3, err := store.GetSubscription(ctx, sub1.ID)
			require.NoError(t, err)
			require.NotNil(t, sub3)

			require.Equal(t, *sub1, *sub3)
		})
	}
}

func TestStoreUpdateSubscription(t *testing.T) {
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
			r2 := sub1
			r2.Url = "new URL"
			sub2, err := store.UpdateSubscription(ctx, r2)
			require.NoError(t, err)
			require.NotNil(t, sub2)
			require.Equal(t, "new URL", sub2.Url)

			// Applying an empty subscription will return a copy
			r3 := r.input.Apply(&models.Subscription{})

			r3.Url = "new URL 2"
			r3.UpdatedAt = time.Now()
			sub3, err := store.UpdateSubscription(ctx, r3)
			require.Error(t, err)
			require.Nil(t, sub3)

			// Update should allow if version not supplied
			r3.Url = "new URL 3"
			r3.UpdatedAt = time.Time{}
			sub4, err := store.UpdateSubscription(ctx, r3)
			require.NoError(t, err)
			require.NotNil(t, sub4)

			// Changing owner should error
			r3.Owner = "new owner"
			sub5, err := store.UpdateSubscription(ctx, r3)
			require.Error(t, err)
			require.Nil(t, sub5)

			sub6, err := store.GetSubscription(ctx, sub1.ID)
			require.NoError(t, err)
			require.NotNil(t, sub6)

			require.Equal(t, *sub4, *sub6)
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
			sub1, err := store.InsertSubscription(ctx, r.input)
			require.NoError(t, err)
			require.NotNil(t, sub1)

			// Ensure mismatched versions return an error
			sub2, err := store.DeleteSubscription(ctx, sub1.ID, sub1.Owner, "a3cg3tcuhk000")
			require.Error(t, err)
			require.Nil(t, sub2)

			sub3, err := store.DeleteSubscription(ctx, sub1.ID, sub1.Owner, sub1.Version())
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
		inserted = []*models.Subscription{}
		cells    = s2.CellUnion{
			s2.CellID(42),
			s2.CellID(84),
			s2.CellID(126),
			s2.CellID(168),
			s2.CellID(200),
		}
		owners = []string{
			"me",
			"my",
			"self",
			"and",
			"i",
			"you",
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
