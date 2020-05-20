package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/stretchr/testify/require"
)

var (
	subscriptionsPool = []struct {
		name  string
		input *ridmodels.Subscription
	}{
		{
			name: "a subscription without startTime and endTime",
			input: &ridmodels.Subscription{
				ID:                dssmodels.ID(uuid.New().String()),
				Owner:             dssmodels.Owner(uuid.New().String()),
				URL:               "https://no/place/like/home",
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription with startTime and endTime",
			input: &ridmodels.Subscription{
				ID:                dssmodels.ID(uuid.New().String()),
				Owner:             dssmodels.Owner(uuid.New().String()),
				URL:               "https://no/place/like/home",
				StartTime:         &startTime,
				EndTime:           &endTime,
				NotificationIndex: 42,
			},
		},
		{
			name: "a subscription with startTime and without endTime",
			input: &ridmodels.Subscription{
				ID:                dssmodels.ID(uuid.New().String()),
				Owner:             dssmodels.Owner(uuid.New().String()),
				URL:               "https://no/place/like/home",
				StartTime:         &startTime,
				NotificationIndex: 42,
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
			},
		},
	}
)

func setUpSubscriptionStore(ctx context.Context, t *testing.T) (*SubscriptionStore, func() error) {
	store, f := setUpStore(ctx, t)
	return store.ISA, f
}

func TestStoreGetSubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpSubscriptionStore(ctx, t)
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
		store, tearDownStore = setUpSubscriptionStore(ctx, t)
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

			// Test changing owner fails
			sub5.Owner = "new bad owner"
			_, err = store.InsertSubscription(ctx, sub5)
			require.Error(t, err)
		})
	}
}

func TestStoreInsertSubscriptionsWithTimes(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpSubscriptionStore(ctx, t)

	defer func() {
		require.NoError(t, tearDownStore())
	}()

	for _, r := range []struct {
		name                string
		updateFromStartTime time.Time
		updateFromEndTime   time.Time
		startTime           time.Time
		endTime             time.Time
		wantErr             string
		wantStartTime       time.Time
		wantEndTime         time.Time
	}{
		{
			name:          "start-time-defaults-to-now",
			endTime:       fakeClock.Now().Add(time.Hour),
			wantStartTime: fakeClock.Now(),
			wantEndTime:   fakeClock.Now().Add(time.Hour),
		},
		{
			name:          "end-time-defaults-to-24h",
			wantStartTime: fakeClock.Now(),
			wantEndTime:   fakeClock.Now().Add(24 * time.Hour),
		},
		{
			name:      "start-time-in-the-past",
			startTime: fakeClock.Now().Add(-6 * time.Minute),
			endTime:   fakeClock.Now().Add(time.Hour),
			wantErr:   "rpc error: code = InvalidArgument desc = subscription time_start must not be in the past",
		},
		{
			name:          "start-time-slightly-in-the-past",
			startTime:     fakeClock.Now().Add(-4 * time.Minute),
			endTime:       fakeClock.Now().Add(time.Hour),
			wantStartTime: fakeClock.Now().Add(-4 * time.Minute),
		},
		{
			name:      "end-time-before-start-time",
			startTime: fakeClock.Now().Add(20 * time.Minute),
			endTime:   fakeClock.Now().Add(10 * time.Minute),
			wantErr:   "rpc error: code = InvalidArgument desc = subscription time_end must be after time_start",
		},
		{
			name:                "updating-keeps-old-times",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(-6 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(6 * time.Hour),
		},
		{
			name:                "changing-start-time-to-past",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			startTime:           fakeClock.Now().Add(-3 * time.Hour),
			wantErr:             "rpc error: code = InvalidArgument desc = subscription time_start must not be in the past",
		},
		{
			name:                "changing-start-time-to-future",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			startTime:           fakeClock.Now().Add(3 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(3 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(6 * time.Hour),
		},
		{
			name:                "changing-end-time-to-future",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			endTime:             fakeClock.Now().Add(3 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(-6 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(3 * time.Hour),
		},
		{
			name:                "changing-end-time-more-than-24h",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			endTime:             fakeClock.Now().Add(24 * time.Hour),
			wantErr:             "rpc error: code = InvalidArgument desc = subscription window exceeds 24 hours",
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			id := dssmodels.ID(uuid.New().String())
			owner := dssmodels.Owner(uuid.New().String())
			var version *dssmodels.Version

			// Insert a pre-existing subscription to simulate updating from something.
			if !r.updateFromStartTime.IsZero() {
				tx, err := store.Begin()
				require.NoError(t, err)
				existing, err := store.pushSubscription(ctx, tx, &ridmodels.Subscription{
					ID:        id,
					Owner:     owner,
					StartTime: &r.updateFromStartTime,
					EndTime:   &r.updateFromEndTime,
				})
				require.NoError(t, err)
				require.NoError(t, tx.Commit())
				version = existing.Version
			}

			s := &ridmodels.Subscription{
				ID:      id,
				Owner:   owner,
				Version: version,
			}
			if !r.startTime.IsZero() {
				s.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				s.EndTime = &r.endTime
			}
			sub, err := store.InsertSubscription(ctx, s)

			if r.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, r.wantErr)
			}

			if !r.wantStartTime.IsZero() {
				require.NotNil(t, sub.StartTime)
				require.Equal(t, r.wantStartTime, *sub.StartTime)
			}
			if !r.wantEndTime.IsZero() {
				require.NotNil(t, sub.EndTime)
				require.Equal(t, r.wantEndTime, *sub.EndTime)
			}
		})
	}
}

func TestStoreInsertTooManySubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpSubscriptionStore(ctx, t)
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
		ret, err := store.InsertSubscription(ctx, makeSubscription([]uint64{42, 43}))
		require.NoError(t, err)
		require.NotNil(t, &ret)
	}

	// Inserting the 11th subscription will fail.
	ret, err := store.InsertSubscription(ctx, makeSubscription([]uint64{42, 43}))
	require.EqualError(t, err, "rpc error: code = ResourceExhausted desc = too many existing subscriptions in this area already")
	require.Nil(t, ret)

	// Inserting a subscription in a different cell will succeed.
	ret, err = store.InsertSubscription(ctx, makeSubscription([]uint64{45}))
	require.NoError(t, err)
	require.NotNil(t, &ret)

	// Inserting a subscription that overlaps with 42 or 43 will fail.
	ret, err = store.InsertSubscription(ctx, makeSubscription([]uint64{7, 43}))
	require.EqualError(t, err, "rpc error: code = ResourceExhausted desc = too many existing subscriptions in this area already")
	require.Nil(t, ret)
}

func TestStoreDeleteSubscription(t *testing.T) {
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpSubscriptionStore(ctx, t)
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
			v, err := dssmodels.VersionFromString("a3cg3tcuhk000")
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
		store, tearDownStore = setUpSubscriptionStore(ctx, t)
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

func TestStoreExpiredSubscription(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpSubscriptionStore(ctx, t)
	defer func() {
		require.NoError(t, tearDownStore())
	}()

	sub := &ridmodels.Subscription{
		ID:    dssmodels.ID(uuid.New().String()),
		Owner: dssmodels.Owner(uuid.New().String()),
		Cells: s2.CellUnion{s2.CellID(42)},
	}

	_, err := store.InsertSubscription(ctx, sub)
	require.NoError(t, err)

	// The subscription's endTime is 24 hours from now.
	fakeClock.Advance(23 * time.Hour)

	// We should still be able to find the subscription by searching and by ID.
	subs, err := store.SearchSubscriptions(ctx, sub.Cells, sub.Owner)
	require.NoError(t, err)
	require.Len(t, subs, 1)

	ret, err := store.GetSubscription(ctx, sub.ID)
	require.NoError(t, err)
	require.NotNil(t, &ret)

	// But now the subscription has expired.
	fakeClock.Advance(2 * time.Hour)

	subs, err = store.SearchSubscriptions(ctx, sub.Cells, sub.Owner)
	require.NoError(t, err)
	require.Len(t, subs, 0)

	ret, err = store.GetSubscription(ctx, sub.ID)
	require.Nil(t, ret)
	require.Error(t, err)
}
