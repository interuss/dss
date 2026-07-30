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
	fakeClock = clockwork.NewFakeClock()
	r := newRepo()
	r.clock = fakeClock
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
