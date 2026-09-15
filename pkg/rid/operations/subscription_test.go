package operations

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/ridv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo/testdata"
	"github.com/interuss/dss/pkg/locality"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

var (
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().Add(-time.Minute)
	endTime   = fakeClock.Now().Add(time.Hour)
)

func newTestContext() context.Context {
	ctx := timestamp.NewContext(context.Background(), fakeClock.Now())
	return locality.NewContext(ctx, "test-locality")
}

type fakeSubscriptionRepo struct {
	subs map[dssmodels.ID]*ridmodels.Subscription
}

func newFakeSubscriptionRepo() *fakeSubscriptionRepo {
	return &fakeSubscriptionRepo{subs: make(map[dssmodels.ID]*ridmodels.Subscription)}
}

func (r *fakeSubscriptionRepo) GetSubscription(_ context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	if sub, ok := r.subs[id]; ok {
		return sub, nil
	}
	return nil, nil
}

func (r *fakeSubscriptionRepo) DeleteSubscription(_ context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	if sub, ok := r.subs[s.ID]; ok {
		delete(r.subs, s.ID)
		return sub, nil
	}
	return nil, nil
}

func (r *fakeSubscriptionRepo) InsertSubscription(_ context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	storedCopy := *s
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	r.subs[s.ID] = &storedCopy
	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (r *fakeSubscriptionRepo) UpdateSubscription(_ context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	storedCopy := *s
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	r.subs[s.ID] = &storedCopy
	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (r *fakeSubscriptionRepo) SearchSubscriptions(_ context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	var subs []*ridmodels.Subscription
	for _, s := range r.subs {
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

func (r *fakeSubscriptionRepo) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	var subs []*ridmodels.Subscription
	res, err := r.SearchSubscriptions(ctx, cells)
	if err != nil {
		return nil, err
	}
	for _, s := range res {
		if s.Owner == owner {
			subs = append(subs, s)
		}
	}
	return subs, nil
}

func (r *fakeSubscriptionRepo) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	subs, err := r.SearchSubscriptions(ctx, cells)
	if err != nil {
		return nil, err
	}
	for i := range subs {
		subs[i].NotificationIndex++
	}
	return subs, nil
}

func (r *fakeSubscriptionRepo) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	maxValue := 0
	subs, err := r.SearchSubscriptionsByOwner(ctx, cells, owner)
	if err != nil {
		return 0, err
	}

	cellMap := make(map[s2.CellID]int)
	for _, s := range subs {
		for _, cid := range s.Cells {
			cellMap[cid]++
			if cellMap[cid] > maxValue {
				maxValue = cellMap[cid]
			}
		}
	}
	return maxValue, nil
}

func (r *fakeSubscriptionRepo) ListExpiredSubscriptions(_ context.Context, _ string, _ time.Time) ([]*ridmodels.Subscription, error) {
	return nil, nil
}

func (r *fakeSubscriptionRepo) CountSubscriptions(_ context.Context) (int64, error) {
	return int64(len(r.subs)), nil
}

func (r *fakeSubscriptionRepo) GetISA(_ context.Context, _ dssmodels.ID, _ bool) (*ridmodels.IdentificationServiceArea, error) {
	panic("not implemented")
}

func (r *fakeSubscriptionRepo) DeleteISA(_ context.Context, _ *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	panic("not implemented")
}

func (r *fakeSubscriptionRepo) InsertISA(_ context.Context, _ *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	panic("not implemented")
}

func (r *fakeSubscriptionRepo) UpdateISA(_ context.Context, _ *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	panic("not implemented")
}

func (r *fakeSubscriptionRepo) SearchISAs(_ context.Context, _ s2.CellUnion, _ *time.Time, _ *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	panic("not implemented")
}

func (r *fakeSubscriptionRepo) ListExpiredISAs(_ context.Context, _ string, _ time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	panic("not implemented")
}

func (r *fakeSubscriptionRepo) CountISAs(_ context.Context) (int64, error) {
	panic("not implemented")
}

func TestInsertSubscriptionsWithTimes(t *testing.T) {
	repo := newFakeSubscriptionRepo()

	for _, r := range []struct {
		name          string
		startTime     time.Time
		endTime       time.Time
		wantErr       stacktrace.ErrorCode
		wantStartTime time.Time
		wantEndTime   time.Time
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
			ctx := newTestContext()
			id := dssmodels.ID(uuid.New().String())
			owner := dssmodels.Owner(uuid.New().String())

			s := &ridmodels.Subscription{
				ID:    id,
				Owner: owner,
				Cells: s2.CellUnion{s2.CellID(17106221850767130624)},
			}
			if !r.startTime.IsZero() {
				s.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				s.EndTime = &r.endTime
			}
			sub, err := InsertSubscription(ctx, repo, s)

			if r.wantErr == stacktrace.ErrorCode(0) {
				require.NoError(t, err)
			} else {
				require.Equal(t, r.wantErr, stacktrace.GetCode(err))
			}

			if !r.wantStartTime.IsZero() {
				require.NotNil(t, sub.StartTime)
				require.Equal(t, r.wantStartTime.UTC().Truncate(time.Microsecond), (*sub.StartTime).UTC().Truncate(time.Microsecond))
			}
			if !r.wantEndTime.IsZero() {
				require.NotNil(t, sub.EndTime)
				require.Equal(t, r.wantEndTime.UTC().Truncate(time.Microsecond), (*sub.EndTime).UTC().Truncate(time.Microsecond))
			}
		})
	}
}

func TestInsertTooManySubscription(t *testing.T) {
	ctx := newTestContext()
	repo := newFakeSubscriptionRepo()

	// Helper function that makes a subscription with a random ID, fixed owner,
	// and provided cellIDs.
	makeSubscription := func(cellIDs []uint64) *ridmodels.Subscription {
		s := &ridmodels.Subscription{
			ID:        dssmodels.ID(uuid.New().String()),
			Owner:     dssmodels.Owner("bob"),
			StartTime: &startTime,
			EndTime:   &endTime,
		}

		s.Cells = make(s2.CellUnion, len(cellIDs))
		for i, id := range cellIDs {
			s.Cells[i] = s2.CellID(id)
		}
		return s
	}

	// We should be able to insert 10 subscriptions without error.
	for i := 0; i < 10; i++ {
		ret, err := InsertSubscription(ctx, repo, makeSubscription([]uint64{12494535901059219456, 12494535866699481088}))
		require.NoError(t, err)
		require.NotNil(t, &ret)
	}

	// Inserting the 11th subscription will fail.
	ret, err := InsertSubscription(ctx, repo, makeSubscription([]uint64{12494535901059219456, 12494535866699481088}))
	require.Equal(t, dsserr.Exhausted, stacktrace.GetCode(err))
	require.Nil(t, ret)

	// Inserting a subscription in a different cell will succeed.
	ret, err = InsertSubscription(ctx, repo, makeSubscription([]uint64{12494535832339742720}))
	require.NoError(t, err)
	require.NotNil(t, &ret)

	// Inserting a subscription that overlaps fail.
	ret, err = InsertSubscription(ctx, repo, makeSubscription([]uint64{12494535935418957824, 12494535866699481088}))
	require.Equal(t, dsserr.Exhausted, stacktrace.GetCode(err))
	require.Nil(t, ret)
}

func TestExecuteInsertSubscriptionValidatesRequest(t *testing.T) {
	for _, r := range []struct {
		name    string
		extents restapi.Volume4D
	}{
		{
			name:    "missing-extents",
			extents: restapi.Volume4D{},
		},
		{
			name: "missing-extents-spatial-volume",
			extents: restapi.Volume4D{
				SpatialVolume: restapi.Volume3D{},
			},
		},
		{
			name: "missing-spatial-volume-footprint",
			extents: restapi.Volume4D{
				SpatialVolume: restapi.Volume3D{
					Footprint: restapi.GeoPolygon{},
				},
			},
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			ctx := newTestContext()
			repo := newFakeSubscriptionRepo()

			req := &restapi.CreateSubscriptionRequest{
				Id: restapi.SubscriptionUUID(uuid.New().String()),
				Body: &restapi.CreateSubscriptionParameters{
					Callbacks: restapi.SubscriptionCallbacks{IdentificationServiceAreaUrl: &testdata.CallbackURL},
					Extents:   r.extents,
				},
				Auth: api.AuthorizationResult{ClientID: &testdata.Owner},
			}

			ret, err := executeInsertSubscription(ctx, repo, req)
			require.Equal(t, dsserr.BadRequest, stacktrace.GetCode(err))
			require.Nil(t, ret)
		})
	}
}
