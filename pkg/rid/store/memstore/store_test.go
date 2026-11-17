package memstore

import (
	"context"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
)

var (
	fakeClock = clockwork.NewFakeClock()
	startTime = fakeClock.Now().UTC().Add(-time.Minute)
	endTime   = fakeClock.Now().UTC().Add(time.Hour)
	writer    = "writer"
)

// setUpStore returns a fresh in-memory repo whose clock is the (reset) package
// fakeClock, so tests can advance time deterministically.
func setUpStore(t *testing.T) *repo {
	t.Helper()
	r := newRepo()
	return r
}

func TestDatabaseEnsuresBeginsBeforeExpires(t *testing.T) {
	ctx := context.Background()
	ctx = timestamp.NewContext(ctx, fakeClock.Now())
	repo := setUpStore(t)

	var (
		begins  = time.Now().UTC()
		expires = begins.Add(-5 * time.Minute)
	)
	_, err := repo.InsertSubscription(ctx, &ridmodels.Subscription{
		ID:                dssmodels.ID(uuid.New().String()),
		Owner:             "me-myself-and-i",
		URL:               "https://no/place/like/home",
		NotificationIndex: 42,
		CellsVolume4D: &dssmodels.CellsVolume4D{
			StartTime: &begins,
			EndTime:   &expires,
		},
	})
	require.Error(t, err)
}

func TestCheckpointRestoreISA(t *testing.T) {
	ctx := context.Background()
	ctx = timestamp.NewContext(ctx, fakeClock.Now())
	repo := setUpStore(t)

	_, err := repo.InsertISA(ctx, serviceArea)
	require.NoError(t, err)

	repo.Checkpoint()

	// Mutate after the checkpoint.
	isa, err := repo.GetISA(ctx, serviceArea.ID, false)
	require.NoError(t, err)
	_, err = repo.DeleteISA(ctx, isa)
	require.NoError(t, err)
	gone, err := repo.GetISA(ctx, serviceArea.ID, false)
	require.NoError(t, err)
	require.Nil(t, gone)

	// Restore brings it back.
	repo.Restore()
	back, err := repo.GetISA(ctx, serviceArea.ID, false)
	require.NoError(t, err)
	require.NotNil(t, back)
}

func TestDeleteExpired(t *testing.T) {
	threshold := fakeClock.Now()
	newStore := func() map[dssmodels.ID]*isaRecord {
		return map[dssmodels.ID]*isaRecord{
			"c":      {ID: "c", EndTime: threshold.Add(-time.Hour)},
			"a":      {ID: "a", EndTime: threshold.Add(-time.Minute)},
			"b":      {ID: "b", EndTime: threshold},
			"active": {ID: "active", EndTime: threshold.Add(time.Nanosecond)},
		}
	}

	testCases := []struct {
		name      string
		limit     int
		deleted   []dssmodels.ID
		remaining []dssmodels.ID
	}{{
		name:      "unlimited",
		limit:     0,
		deleted:   []dssmodels.ID{"a", "b", "c"},
		remaining: []dssmodels.ID{"active"},
	}, {
		name:      "limit keeps the smallest IDs",
		limit:     2,
		deleted:   []dssmodels.ID{"a", "b"},
		remaining: []dssmodels.ID{"active", "c"},
	}, {
		name:      "limit above the number of expired records",
		limit:     10,
		deleted:   []dssmodels.ID{"a", "b", "c"},
		remaining: []dssmodels.ID{"active"},
	}}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := newStore()
			require.Equal(t, testCase.deleted, deleteExpired(store, threshold, testCase.limit))
			require.ElementsMatch(t, testCase.remaining, slices.Collect(maps.Keys(store)))
		})
	}
}

func TestListExpired(t *testing.T) {
	threshold := fakeClock.Now()
	store := map[dssmodels.ID]*isaRecord{
		"c":      {ID: "c", EndTime: threshold.Add(-time.Hour)},
		"a":      {ID: "a", EndTime: threshold.Add(-time.Minute)},
		"b":      {ID: "b", EndTime: threshold},
		"active": {ID: "active", EndTime: threshold.Add(time.Nanosecond)},
	}

	require.Equal(t, []dssmodels.ID{"a", "b", "c"}, listExpired(store, threshold))
	require.Len(t, store, 4, "listing must not delete")
}

func TestCheckpointIsolatesNotificationIndex(t *testing.T) {
	ctx := context.Background()
	ctx = timestamp.NewContext(ctx, fakeClock.Now())
	repo := setUpStore(t)

	sub, err := repo.InsertSubscription(ctx, subscriptionsPool[0].input)
	require.NoError(t, err)

	repo.Checkpoint()

	// In-place notification-index bump must not leak into the checkpoint.
	updated, err := repo.UpdateNotificationIdxsInCells(ctx, sub.Cells)
	require.NoError(t, err)
	require.Len(t, updated, 1)
	require.Equal(t, sub.NotificationIndex+1, updated[0].NotificationIndex)

	repo.Restore()
	restored, err := repo.GetSubscription(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, sub.NotificationIndex, restored.NotificationIndex)
}
