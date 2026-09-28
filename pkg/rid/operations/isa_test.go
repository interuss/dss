package operations

import (
	"context"
	"testing"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/require"
)

func deleteISA(ctx context.Context, repo repos.Repository, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ISAResult, error) {
	req := NewDeleteISAPayload(id, owner, version)
	ret, err := executeDeleteISA(ctx, repo, req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return nil, nil
	}
	return ret.(*ISAResult), nil
}

func TestDeleteISA(t *testing.T) {
	ctx := newTestContext()
	repo := newFakeSubscriptionRepo()

	insertedSubscriptions := make([]*ridmodels.Subscription, 0, 2)
	for range 2 {
		s, err := insertSubscription(ctx, repo, &ridmodels.Subscription{
			ID:    dssmodels.ID(uuid.New().String()),
			Owner: "owner",
			URL:   "https://no/place/like/home",
			CellsVolume4D: &dssmodels.CellsVolume4D{
				StartTime: &startTime,
				EndTime:   &endTime,
				Cells:     s2.CellUnion{12494535935418957824},
			},
		})
		require.NoError(t, err)
		insertedSubscriptions = append(insertedSubscriptions, s)
	}
	for _, s := range insertedSubscriptions {
		require.Equal(t, 0, s.NotificationIndex)
	}

	// Insert the ISA.
	serviceArea := &ridmodels.IdentificationServiceArea{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     dssmodels.Owner(uuid.New().String()),
		URL:       "https://no/place/like/home/for/flights",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells:     s2.CellUnion{12494535935418957824},
	}
	insertSubs, err := repo.UpdateNotificationIdxsInCells(ctx, serviceArea.Cells)
	require.NoError(t, err)
	require.Len(t, insertSubs, len(insertedSubscriptions))
	for _, s := range insertSubs {
		require.Equal(t, 1, s.NotificationIndex)
	}
	isa, err := repo.InsertISA(ctx, serviceArea)
	require.NoError(t, err)
	require.NotNil(t, isa)

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
