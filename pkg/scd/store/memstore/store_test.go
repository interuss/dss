package memstore

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/stretchr/testify/require"
)

var (
	manager = dssmodels.Manager("unittest")

	constraintId        = dssmodels.ID("00000185-e36d-40be-8d38-beca6ca31000")
	subscriptionId      = dssmodels.ID("00000185-e36d-40be-8d38-beca6ca31001")
	operationalIntentId = dssmodels.ID("00000185-e36d-40be-8d38-beca6ca31002")

	cells = s2.CellUnion{
		s2.CellID(int64(8768904281496485888)),
		s2.CellID(int64(8768904178417270784)),
	}

	startTime = time.Date(2024, time.August, 14, 15, 48, 36, 0, time.UTC)
	endTime   = startTime.Add(time.Hour)
	writeTime = time.Date(2024, time.August, 1, 0, 0, 0, 0, time.UTC)

	altLow, altHigh float32 = 84, 169
)

// setUpStore returns a fresh in-memory repo.
func setUpStore(t *testing.T) *repo {
	t.Helper()
	return newRepo()
}

// writeCtx returns a context carrying a deterministic write timestamp so that
// updated_at is controlled in tests.
func writeCtx() context.Context {
	return timestamp.WithTimestamp(context.Background(), writeTime)
}

func sampleConstraint() *scdmodels.Constraint {
	return &scdmodels.Constraint{
		ID:            constraintId,
		Manager:       manager,
		Version:       1,
		StartTime:     &startTime,
		EndTime:       &endTime,
		USSBaseURL:    "https://dummy.uss",
		AltitudeLower: &altLow,
		AltitudeUpper: &altHigh,
		Cells:         cells,
	}
}

func sampleSubscription() *scdmodels.Subscription {
	return &scdmodels.Subscription{
		ID:                          subscriptionId,
		Manager:                     manager,
		NotificationIndex:           1,
		USSBaseURL:                  "https://dummy.uss",
		NotifyForOperationalIntents: true,
		NotifyForConstraints:        true,
		StartTime:                   &startTime,
		EndTime:                     &endTime,
		Cells:                       cells,
	}
}

func sampleOperationalIntent() *scdmodels.OperationalIntent {
	sid := subscriptionId
	return &scdmodels.OperationalIntent{
		ID:             operationalIntentId,
		Manager:        manager,
		Version:        1,
		State:          scdmodels.OperationalIntentStateAccepted,
		StartTime:      &startTime,
		EndTime:        &endTime,
		USSBaseURL:     "https://dummy.uss",
		SubscriptionID: &sid,
		AltitudeLower:  &altLow,
		AltitudeUpper:  &altHigh,
		Cells:          cells,
	}
}

func sampleAvailability() *scdmodels.UssAvailabilityStatus {
	return &scdmodels.UssAvailabilityStatus{
		Uss:          manager,
		Availability: scdmodels.UssAvailabilityStateNormal,
	}
}

// volume4D builds a Volume4D whose footprint covers the provided cells.
func volume4D(cu s2.CellUnion, start, end *time.Time, altLo, altHi *float32) *dssmodels.Volume4D {
	return &dssmodels.Volume4D{
		StartTime: start,
		EndTime:   end,
		SpatialVolume: &dssmodels.Volume3D{
			AltitudeLo: altLo,
			AltitudeHi: altHi,
			Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
				return cu, nil
			}),
		},
	}
}

func TestCheckpointRestoreRoundTrip(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)

	_, err := r.UpsertConstraint(ctx, sampleConstraint())
	require.NoError(t, err)
	_, err = r.UpsertSubscription(ctx, sampleSubscription())
	require.NoError(t, err)
	_, err = r.UpsertOperationalIntent(ctx, sampleOperationalIntent())
	require.NoError(t, err)
	_, err = r.UpsertUssAvailability(ctx, sampleAvailability())
	require.NoError(t, err)

	cp := r.Checkpoint()

	// Mutate after the checkpoint.
	require.NoError(t, r.DeleteConstraint(ctx, constraintId))
	require.NoError(t, r.DeleteSubscription(ctx, subscriptionId))
	require.NoError(t, r.DeleteOperationalIntent(ctx, operationalIntentId))

	// Restore brings everything back.
	require.NoError(t, r.Restore(cp))

	con, err := r.GetConstraint(ctx, constraintId)
	require.NoError(t, err)
	require.NotNil(t, con)
	sub, err := r.GetSubscription(ctx, subscriptionId)
	require.NoError(t, err)
	require.NotNil(t, sub)
	oi, err := r.GetOperationalIntent(ctx, operationalIntentId)
	require.NoError(t, err)
	require.NotNil(t, oi)
}

func TestCheckpointIsolatesNotificationIndex(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)

	sub, err := r.UpsertSubscription(ctx, sampleSubscription())
	require.NoError(t, err)

	cp := r.Checkpoint()

	// In-place notification-index bump must not leak into the checkpoint.
	bumped, err := r.IncrementNotificationIndicesForOperationalIntents(ctx, volume4D(cells, nil, nil, nil, nil))
	require.NoError(t, err)
	require.Len(t, bumped, 1)
	require.Equal(t, sub.NotificationIndex+1, bumped[0].NotificationIndex)

	require.NoError(t, r.Restore(cp))
	restored, err := r.GetSubscription(ctx, subscriptionId)
	require.NoError(t, err)
	require.Equal(t, sub.NotificationIndex, restored.NotificationIndex)
}

func TestRestoreInvalidType(t *testing.T) {
	require.Error(t, setUpStore(t).Restore("not a checkpoint"))
}
