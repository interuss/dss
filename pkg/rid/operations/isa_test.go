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

	// Can't delete with a stale version.
	_, err = deleteISA(ctx, repo, isa.ID, isa.Owner, dssmodels.VersionFromTime(time.Now().Add(-time.Hour)))
	require.Equal(t, dsserr.VersionMismatch, stacktrace.GetCode(err))

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
