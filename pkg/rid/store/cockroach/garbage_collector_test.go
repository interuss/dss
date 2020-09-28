package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/dpjacques/clockwork"
	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/stretchr/testify/require"
)

func TestDeleteExpiredISAs(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	fakeClock := clockwork.NewFakeClockAt(time.Now())

	// Insert ISA with endtime to 30 minutes ago
	startTime := fakeClock.Now().Add(-1 * time.Hour)
	endTime := fakeClock.Now().Add(-30 * time.Minute)
	serviceArea := &ridmodels.IdentificationServiceArea{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     dssmodels.Owner(uuid.New().String()),
		URL:       "https://no/place/like/home/for/flights",
		StartTime: &startTime,
		EndTime:   &endTime,
		Writer:    writer,
		Cells: s2.CellUnion{
			s2.CellID(uint64(overflow)),
			s2.CellID(17106221850767130624),
		},
	}
	saOut, err := repo.InsertISA(ctx, serviceArea)
	require.NoError(t, err)
	require.NotNil(t, saOut)

	// A get should work even if it is expired.
	ret, err := repo.GetISA(ctx, serviceArea.ID)
	require.NoError(t, err)
	require.NotNil(t, ret)

	gc := NewGarbageCollector(repo, writer)
	gc.DeleteExpiredRecords(ctx)

	// A get should work even if it is expired.
	ret, err = repo.GetISA(ctx, serviceArea.ID)
	require.NoError(t, err)
	require.Nil(t, ret)
}

func TestDeleteExpiredSubscriptions(t *testing.T) {
	ctx := context.Background()
	store, tearDownStore := setUpStore(ctx, t)
	defer tearDownStore()

	repo, err := store.Interact(ctx)
	require.NoError(t, err)

	fakeClock := clockwork.NewFakeClockAt(time.Now())

	// Insert ISA with endtime to 30 minutes ago
	startTime := fakeClock.Now().Add(-1 * time.Hour)
	endTime := fakeClock.Now().Add(-30 * time.Minute)
	subscription := &ridmodels.Subscription{
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
	}

	subOut, err := repo.InsertSubscription(ctx, subscription)
	require.NoError(t, err)
	require.NotNil(t, subOut)

	// A get should work even if it is expired.
	ret, err := repo.GetSubscription(ctx, subscription.ID)
	require.NoError(t, err)
	require.NotNil(t, ret)

	gc := NewGarbageCollector(repo, writer)
	gc.DeleteExpiredRecords(ctx)

	// A get should work even if it is expired.
	ret, err = repo.GetSubscription(ctx, subscription.ID)
	require.NoError(t, err)
	require.Nil(t, ret)
}
