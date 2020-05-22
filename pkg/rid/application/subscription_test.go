package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dpjacques/clockwork"
	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/stretchr/testify/require"
)

var (
	fakeClock = clockwork.NewFakeClock()
)

func setUpSubApp() SubscriptionApp {
	return &app{
		Subscription: &subscriptionStore{
			subs: make(map[dssmodels.ID]*ridmodels.Subscription),
		},
		clock: fakeClock,
	}
}

type subscriptionStore struct {
	subs map[dssmodels.ID]*ridmodels.Subscription
}

func (store *subscriptionStore) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	if sub, ok := store.subs[id]; ok {
		return sub, nil
	}
	return nil, sql.ErrNoRows
}

// DeleteSubscription deletes the Subscription identified by "id" and owned by "owner".
// Returns the delete Subscription and all IdentificationServiceAreas affected by the delete.
func (store *subscriptionStore) DeleteSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	if sub, ok := store.subs[s.ID]; ok {
		delete(store.subs, s.ID)
		return sub, nil
	}
	return nil, sql.ErrNoRows
}

// InsertSubscription inserts or updates an Subscription.
func (store *subscriptionStore) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	storedCopy := *s
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	store.subs[s.ID] = &storedCopy

	returnedCopy := storedCopy
	return &returnedCopy, nil
}

// Update
func (store *subscriptionStore) UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	return nil, errors.New("not implemented")
}

// SearchSubscriptionsByOwner returns all IdentificationServiceAreas ownded by "owner" in "cells".
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

func (store *subscriptionStore) updateNotificationIdxs(ctx context.Context, cells s2.CellUnion) []*ridmodels.Subscription {
	var ret []*ridmodels.Subscription
	subs, _ := store.SearchSubscriptions(ctx, cells)
	for _, s := range subs {
		s.NotificationIndex++
		s, _ = store.InsertSubscription(ctx, s)
		ret = append(ret, s)
	}
	return ret
}

// SearchSubscriptions returns all IdentificationServiceAreas ownded by "owner" in "cells".
func (store *subscriptionStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	var subs []*ridmodels.Subscription

	for _, s := range store.subs {
		if s.Cells.Intersects(cells) {
			subs = append(subs, s)
		}
	}
	return subs, nil
}

func TestBadOwner(t *testing.T) {
	ctx := context.Background()
	app := setUpSubApp()

	sub := &ridmodels.Subscription{
		ID:    dssmodels.ID(uuid.New().String()),
		Owner: "orig Owner",
		Cells: s2.CellUnion{s2.CellID(42)},
	}

	sub, err := app.InsertSubscription(ctx, sub)
	require.NoError(t, err)
	// Test changing owner fails
	sub.Owner = "new bad owner"
	_, err = app.InsertSubscription(ctx, sub)
	require.EqualError(t, err, fmt.Sprintf("rpc error: code = PermissionDenied desc = s is owned by orig Owner"))
}

func TestInsertSubscriptionsWithTimes(t *testing.T) {
	ctx := context.Background()
	app := setUpSubApp()

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
				existing, err := app.InsertSubscription(ctx, &ridmodels.Subscription{
					ID:        id,
					Owner:     owner,
					StartTime: &r.updateFromStartTime,
					EndTime:   &r.updateFromEndTime,
				})
				require.NoError(t, err)
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
			sub, err := app.InsertSubscription(ctx, s)

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
