package operations

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/require"
)

func TestInsertISA(t *testing.T) {
	ctx := newTestContext()
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
			name:    "missing-end-time",
			wantErr: dsserr.BadRequest,
		},
		{
			name:          "start-time-defaults-to-now",
			endTime:       fakeClock.Now().Add(time.Hour),
			wantStartTime: fakeClock.Now(),
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
			sa := &ridmodels.IdentificationServiceArea{
				ID:    dssmodels.ID(uuid.New().String()),
				Owner: dssmodels.Owner(uuid.New().String()),
				Cells: s2.CellUnion{12494535935418957824},
			}
			if !r.startTime.IsZero() {
				sa.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				sa.EndTime = &r.endTime
			}
			result, err := insertISA(ctx, repo, sa)

			if r.wantErr == stacktrace.ErrorCode(0) {
				require.NoError(t, err)
			} else {
				require.Equal(t, r.wantErr, stacktrace.GetCode(err))
			}

			if !r.wantStartTime.IsZero() {
				require.NotNil(t, result.ISA.StartTime)
				// time.Time times are represented with loc==nil. The nil location means UTC.
				// for test equality, it has to be explicitly converted to UTC.
				// similar issue: https://github.com/golang/go/issues/19486
				require.Equal(t, r.wantStartTime.UTC().Truncate(time.Microsecond), (*result.ISA.StartTime).UTC().Truncate(time.Microsecond))
			}
			if !r.wantEndTime.IsZero() {
				require.NotNil(t, result.ISA.EndTime)
				require.Equal(t, r.wantEndTime.UTC().Truncate(time.Microsecond), (*result.ISA.EndTime).UTC().Truncate(time.Microsecond))
			}
		})
	}
}

func TestDeleteISA(t *testing.T) {
	ctx := newTestContext()
	repo := newFakeSubscriptionRepo()

	insertedSubscriptions := make([]*ridmodels.Subscription, 0, 2)
	for range 2 {
		s, err := InsertSubscription(ctx, repo, &ridmodels.Subscription{
			ID:        dssmodels.ID(uuid.New().String()),
			Owner:     "owner",
			URL:       "https://no/place/like/home",
			StartTime: &startTime,
			EndTime:   &endTime,
			Cells:     s2.CellUnion{12494535935418957824},
		})
		require.NoError(t, err)
		insertedSubscriptions = append(insertedSubscriptions, s)
	}
	for _, s := range insertedSubscriptions {
		require.Equal(t, 0, s.NotificationIndex)
	}

	// Insert the ISA.
	insertResult, err := insertISA(ctx, repo, &ridmodels.IdentificationServiceArea{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     dssmodels.Owner(uuid.New().String()),
		URL:       "https://no/place/like/home/for/flights",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells:     s2.CellUnion{12494535935418957824},
	})
	require.NoError(t, err)
	require.NotNil(t, insertResult)
	require.Len(t, insertResult.Subscriptions, len(insertedSubscriptions))
	for _, s := range insertResult.Subscriptions {
		require.Equal(t, 1, s.NotificationIndex)
	}

	isa := insertResult.ISA

	// Can't delete with different owner.
	_, err = deleteISA(ctx, repo, isa.ID, "bad-owner", isa.Version)
	require.Equal(t, dsserr.PermissionDenied, stacktrace.GetCode(err))

	deleteResult, err := deleteISA(ctx, repo, isa.ID, isa.Owner, isa.Version)
	require.NoError(t, err)
	require.Equal(t, isa, deleteResult.ISA)
	require.Len(t, deleteResult.Subscriptions, len(insertedSubscriptions))
	for _, s := range deleteResult.Subscriptions {
		require.Equal(t, 2, s.NotificationIndex)
	}

	// Deleting again fails since it no longer exists.
	_, err = deleteISA(ctx, repo, isa.ID, isa.Owner, isa.Version)
	require.Equal(t, dsserr.NotFound, stacktrace.GetCode(err))
}

func TestUpdateISA(t *testing.T) {
	ctx := newTestContext()
	repo := newFakeSubscriptionRepo()

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
	} {
		t.Run(r.name, func(t *testing.T) {
			id := dssmodels.ID(uuid.New().String())
			owner := dssmodels.Owner(uuid.New().String())

			// Insert a pre-existing ISA to simulate updating from something.
			existing, err := repo.InsertISA(ctx, &ridmodels.IdentificationServiceArea{
				ID:        id,
				Owner:     owner,
				StartTime: &r.updateFromStartTime,
				EndTime:   &r.updateFromEndTime,
				Cells:     s2.CellUnion{12494535935418957824},
			})
			require.NoError(t, err)

			sa := &ridmodels.IdentificationServiceArea{
				ID:      id,
				Owner:   owner,
				Version: existing.Version,
				Cells:   s2.CellUnion{12494535935418957824},
			}
			if !r.startTime.IsZero() {
				sa.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				sa.EndTime = &r.endTime
			}
			result, err := updateISA(ctx, repo, sa)

			if r.wantErr == stacktrace.ErrorCode(0) {
				require.NoError(t, err)
			} else {
				require.Equal(t, r.wantErr, stacktrace.GetCode(err))
			}

			if !r.wantStartTime.IsZero() {
				require.NotNil(t, result.ISA.StartTime)
				require.Equal(t, r.wantStartTime.UTC().Truncate(time.Microsecond), (*result.ISA.StartTime).UTC().Truncate(time.Microsecond))
			}
			if !r.wantEndTime.IsZero() {
				require.NotNil(t, result.ISA.EndTime)
				require.Equal(t, r.wantEndTime.UTC().Truncate(time.Microsecond), (*result.ISA.EndTime).UTC().Truncate(time.Microsecond))
			}
		})
	}
}

func TestISAUpdateIdxCells(t *testing.T) {
	ctx := newTestContext()
	repo := newFakeSubscriptionRepo()

	// ensure that when we do an update, nothing in the s2 library joins multiple
	// cells together at a lower level.

	// These 4 cells are fully encompassed by the parent cell, meaning the s2
	// library might try to Normalize (this is the name of the function) the Union
	// into a single cell. We don't support this currently, so let's make sure
	// this doesn't happen.
	insertResult, err := insertISA(ctx, repo, &ridmodels.IdentificationServiceArea{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     "owner",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells:     s2.CellUnion{17106221850767130624, 17106221885126868992, 17106221919486607360},
	})
	require.NoError(t, err)
	require.NotNil(t, insertResult)
	isa := insertResult.ISA

	// Now insert 2 subs, one overlaps with the original isa, and the second, overlaps
	// with the soon to be new version of the isa. both should increase their
	// notification index.
	_, err = InsertSubscription(ctx, repo, &ridmodels.Subscription{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     "owner",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells:     s2.CellUnion{17106221850767130624, 17106221919486607360},
	})
	require.NoError(t, err)

	_, err = InsertSubscription(ctx, repo, &ridmodels.Subscription{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     "owner",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells:     s2.CellUnion{17106221953846345728},
	})
	require.NoError(t, err)

	isa.Cells = s2.CellUnion{17106221953846345728}

	updateResult, err := updateISA(ctx, repo, isa)
	require.NoError(t, err)
	require.NotNil(t, updateResult)
	require.Len(t, updateResult.Subscriptions, 2)
	for _, sub := range updateResult.Subscriptions {
		require.Equal(t, 1, sub.NotificationIndex)
	}

	isas, err := repo.SearchISAs(ctx, updateResult.ISA.Cells, &startTime, nil)
	require.NoError(t, err)
	require.NotNil(t, isas)
	require.Len(t, isas, 1)
}
