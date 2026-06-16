package memstore

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
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
		StartTime:         &begins,
		EndTime:           &expires,
	})
	require.Error(t, err)
}

func TestCheckpointRestoreISA(t *testing.T) {
	ctx := context.Background()
	repo := setUpStore(t)

	_, err := repo.InsertISA(ctx, serviceArea)
	require.NoError(t, err)

	cp := repo.Checkpoint()

	// Mutate after the checkpoint.
	isa, err := repo.GetISA(ctx, serviceArea.ID, false)
	require.NoError(t, err)
	_, err = repo.DeleteISA(ctx, isa)
	require.NoError(t, err)
	gone, err := repo.GetISA(ctx, serviceArea.ID, false)
	require.NoError(t, err)
	require.Nil(t, gone)

	// Restore brings it back.
	require.NoError(t, repo.Restore(cp))
	back, err := repo.GetISA(ctx, serviceArea.ID, false)
	require.NoError(t, err)
	require.NotNil(t, back)
}

func TestCheckpointIsolatesNotificationIndex(t *testing.T) {
	ctx := context.Background()
	repo := setUpStore(t)

	sub, err := repo.InsertSubscription(ctx, subscriptionsPool[0].input)
	require.NoError(t, err)

	cp := repo.Checkpoint()

	// In-place notification-index bump must not leak into the checkpoint.
	updated, err := repo.UpdateNotificationIdxsInCells(ctx, sub.Cells)
	require.NoError(t, err)
	require.Len(t, updated, 1)
	require.Equal(t, sub.NotificationIndex+1, updated[0].NotificationIndex)

	require.NoError(t, repo.Restore(cp))
	restored, err := repo.GetSubscription(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, sub.NotificationIndex, restored.NotificationIndex)
}

func TestRestoreInvalidType(t *testing.T) {
	require.Error(t, setUpStore(t).Restore("not a checkpoint"))
}
