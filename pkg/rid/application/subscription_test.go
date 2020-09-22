package application

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	// Ensure the struct conforms to the interface
	_                 SubscriptionApp = &app{}
	subscriptionsPool                 = []struct {
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

func setUpSubApp(ctx context.Context, t *testing.T) (*app, func()) {
	l := zap.L()
	transactor, cleanup := setUpStore(ctx, t, l)
	return NewFromTransactor(transactor, l).(*app), cleanup
}

type subscriptionStore struct {
	subs map[dssmodels.ID]*ridmodels.Subscription
}

func (store *subscriptionStore) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	if sub, ok := store.subs[id]; ok {
		return sub, nil
	}
	return nil, nil
}

// DeleteSubscription deletes the Subscription identified by "id" and owned by "owner".
// Returns the delete Subscription and all IdentificationServiceAreas affected by the delete.
func (store *subscriptionStore) DeleteSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	if sub, ok := store.subs[s.ID]; ok {
		delete(store.subs, s.ID)
		return sub, nil
	}
	return nil, nil
}

func (store *subscriptionStore) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	storedCopy := *s
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	store.subs[s.ID] = &storedCopy

	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (store *subscriptionStore) UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	storedCopy := *s
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	store.subs[s.ID] = &storedCopy

	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (store *subscriptionStore) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	var subs []*ridmodels.Subscription

	res, _ := store.SearchSubscriptions(ctx, cells)
	for _, s := range res {
		if s.Owner == owner {
			subs = append(subs, s)
		}
	}
	return subs, nil
}

func (store *subscriptionStore) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	subs, _ := store.SearchSubscriptions(ctx, cells)
	for i := range subs {
		subs[i].NotificationIndex++
	}
	return subs, nil
}

func (store *subscriptionStore) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	max := 0
	subs, _ := store.SearchSubscriptionsByOwner(ctx, cells, owner)

	cellMap := make(map[s2.CellID]int)
	for _, s := range subs {
		for _, cid := range s.Cells {
			if _, ok := cellMap[cid]; !ok {
				cellMap[cid] = 1
			} else {
				cellMap[cid]++
			}
			if cellMap[cid] > max {
				max = cellMap[cid]
			}
		}
	}
	return max, nil
}

func (store *subscriptionStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	var subs []*ridmodels.Subscription
	for _, s := range store.subs {
		// Don't call Intersects, since that's smarter code than we implement in the DB.
		appended := false
		for _, c1 := range s.Cells {
			for _, c2 := range cells {
				if c1 == c2 {
					subs = append(subs, s)
					appended = true
					break
				}
			}
			if appended {
				break
			}
		}
	}
	return subs, nil
}

func (store *subscriptionStore) ListExpiredSubscriptions(ctx context.Context, writer *string) ([]*ridmodels.Subscription, error) {
	return make([]*ridmodels.Subscription, 0), nil
}

func TestBadOwner(t *testing.T) {
	ctx := context.Background()
	app, cleanup := setUpSubApp(ctx, t)
	defer cleanup()

	sub := &ridmodels.Subscription{
		ID:    dssmodels.ID(uuid.New().String()),
		Owner: "orig Owner",
		Cells: s2.CellUnion{s2.CellID(17106221850767130624)},
	}

	sub, err := app.InsertSubscription(ctx, sub)
	require.NoError(t, err)
	// Test changing owner fails
	sub.Owner = "new bad owner"
	_, err = app.UpdateSubscription(ctx, sub)
	require.Equal(t, stacktrace.GetCode(err), dsserr.PermissionDenied)
}

func TestSubscriptionUpdateCells(t *testing.T) {
	ctx := context.Background()
	owner := dssmodels.Owner("owner")
	app, cleanup := setUpSubApp(ctx, t)
	defer cleanup()

	// ensure that when we do an update, nothing in the s2 library joins multiple
	// cells together at a lower level.

	// These 4 cells are fully encompassed by the parent cell, meaning the s2
	// library might try to Normalize (this is the name of the function) the Union
	// into a single cell. We don't support this currently, so let's make sure
	// this doesn't happen.
	sub, err := app.InsertSubscription(ctx, &ridmodels.Subscription{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     owner,
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells:     s2.CellUnion{17106221850767130624, 17106221885126868992, 17106221919486607360},
	})

	require.NoError(t, err)
	require.NotNil(t, sub)

	sub.Cells = s2.CellUnion{17106221953846345728}

	sub, err = app.UpdateSubscription(ctx, sub)
	require.NoError(t, err)
	require.NotNil(t, sub)

	subs, err := app.SearchSubscriptionsByOwner(ctx, sub.Cells, owner)
	require.NoError(t, err)
	require.NotNil(t, subs)
	require.Len(t, subs, 1)
}

func TestInsertSubscriptionsWithTimes(t *testing.T) {
	ctx := context.Background()
	app, cleanup := setUpSubApp(ctx, t)
	defer cleanup()

	for _, r := range []struct {
		name                string
		updateFromStartTime time.Time
		updateFromEndTime   time.Time
		startTime           time.Time
		endTime             time.Time
		wantErr             stacktrace.ErrorCode
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
			wantErr:   dsserr.BadRequest,
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
			wantErr:   dsserr.BadRequest,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			id := dssmodels.ID(uuid.New().String())
			owner := dssmodels.Owner(uuid.New().String())
			var version *dssmodels.Version

			s := &ridmodels.Subscription{
				ID:      id,
				Owner:   owner,
				Version: version,
				Cells:   s2.CellUnion{s2.CellID(17106221850767130624)},
			}
			if !r.startTime.IsZero() {
				s.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				s.EndTime = &r.endTime
			}
			sub, err := app.InsertSubscription(ctx, s)

			if r.wantErr == stacktrace.ErrorCode(0) {
				require.NoError(t, err)
			} else {
				require.Equal(t, stacktrace.GetCode(err), r.wantErr)
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

func TestUpdateSubscriptionsWithTimes(t *testing.T) {
	ctx := context.Background()
	app, cleanup := setUpSubApp(ctx, t)
	defer cleanup()

	for _, r := range []struct {
		name                string
		updateFromStartTime time.Time
		updateFromEndTime   time.Time
		startTime           time.Time
		endTime             time.Time
		wantErr             stacktrace.ErrorCode
		wantStartTime       time.Time
		wantEndTime         time.Time
	}{
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
			wantErr:             dsserr.BadRequest,
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
			wantErr:             dsserr.BadRequest,
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			var (
				id      = dssmodels.ID(uuid.New().String())
				owner   = dssmodels.Owner(uuid.New().String())
				version *dssmodels.Version
			)

			repo, err := app.Store.Interact(ctx)
			require.NoError(t, err)

			// Insert a pre-existing subscription to simulate updating from something.
			existing, err := repo.InsertSubscription(ctx, &ridmodels.Subscription{
				ID:        id,
				Owner:     owner,
				StartTime: &r.updateFromStartTime,
				EndTime:   &r.updateFromEndTime,
				Cells:     s2.CellUnion{s2.CellID(17106221850767130624)},
			})
			require.NoError(t, err)
			version = existing.Version

			s := &ridmodels.Subscription{
				ID:      id,
				Owner:   owner,
				Version: version,
				Cells:   s2.CellUnion{s2.CellID(17106221850767130624)},
			}
			if !r.startTime.IsZero() {
				s.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				s.EndTime = &r.endTime
			}
			sub, err := app.UpdateSubscription(ctx, s)

			if r.wantErr == stacktrace.ErrorCode(0) {
				require.NoError(t, err)
			} else {
				require.Equal(t, stacktrace.GetCode(err), r.wantErr)
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

func TestInsertTooManySubscription(t *testing.T) {
	var (
		ctx          = context.Background()
		app, cleanup = setUpSubApp(ctx, t)
	)
	defer cleanup()
	// Helper function that makes a subscription with a random ID, fixed owner,
	// and provided cellIDs.
	makeSubscription := func(cellIDs []uint64) *ridmodels.Subscription {
		s := &ridmodels.Subscription{
			ID:        dssmodels.ID(uuid.New().String()),
			Owner:     dssmodels.Owner("bob"),
			StartTime: &startTime,
			EndTime:   &endTime,
			Cells:     s2.CellUnion{s2.CellID(17106221850767130624)},
		}

		s.Cells = make(s2.CellUnion, len(cellIDs))
		for i, id := range cellIDs {
			s.Cells[i] = s2.CellID(id)
		}
		return s
	}

	// We should be able to insert 10 subscriptions without error.
	for i := 0; i < 10; i++ {
		ret, err := app.InsertSubscription(ctx, makeSubscription([]uint64{12494535901059219456, 12494535866699481088}))
		require.NoError(t, err)
		require.NotNil(t, &ret)
	}

	// Inserting the 11th subscription will fail.
	ret, err := app.InsertSubscription(ctx, makeSubscription([]uint64{12494535901059219456, 12494535866699481088}))
	require.Equal(t, stacktrace.GetCode(err), dsserr.Exhausted)
	require.Nil(t, ret)

	// Inserting a subscription in a different cell will succeed.
	ret, err = app.InsertSubscription(ctx, makeSubscription([]uint64{12494535832339742720}))
	require.NoError(t, err)
	require.NotNil(t, &ret)

	// Inserting a subscription that overlaps fail.
	ret, err = app.InsertSubscription(ctx, makeSubscription([]uint64{12494535935418957824, 12494535866699481088}))
	require.Equal(t, stacktrace.GetCode(err), dsserr.Exhausted)
	require.Nil(t, ret)
}
