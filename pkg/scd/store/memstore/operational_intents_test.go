package memstore

import (
	"testing"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/stretchr/testify/require"
)

func TestOperationalIntentUpsertGetDelete(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)

	got, err := r.UpsertOperationalIntent(ctx, sampleOperationalIntent())
	require.NoError(t, err)
	require.Equal(t, operationalIntentId, got.ID)
	require.Equal(t, scdmodels.OperationalIntentStateAccepted, got.State)
	require.NotEmpty(t, got.OVN)
	// No availability stored yet: defaults to Unknown.
	require.Equal(t, scdmodels.UssAvailabilityStateUnknown, got.UssAvailability)

	count, err := r.CountOperationalIntents(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	require.NoError(t, r.DeleteOperationalIntent(ctx, operationalIntentId))
	gone, err := r.GetOperationalIntent(ctx, operationalIntentId)
	require.NoError(t, err)
	require.Nil(t, gone)
}

func TestOperationalIntentGetMissingReturnsNil(t *testing.T) {
	r := setUpStore(t)
	got, err := r.GetOperationalIntent(writeCtx(), operationalIntentId)
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestOperationalIntentDeleteMissingErrors(t *testing.T) {
	r := setUpStore(t)
	require.Error(t, r.DeleteOperationalIntent(writeCtx(), operationalIntentId))
}

func TestOperationalIntentUssAvailabilityAttached(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)
	_, err := r.UpsertUssAvailability(ctx, sampleAvailability())
	require.NoError(t, err)
	_, err = r.UpsertOperationalIntent(ctx, sampleOperationalIntent())
	require.NoError(t, err)

	got, err := r.GetOperationalIntent(ctx, operationalIntentId)
	require.NoError(t, err)
	require.Equal(t, scdmodels.UssAvailabilityStateNormal, got.UssAvailability)
}

func TestSearchOperationalIntents(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)
	_, err := r.UpsertOperationalIntent(ctx, sampleOperationalIntent())
	require.NoError(t, err)

	res, err := r.SearchOperationalIntents(ctx, volume4D(cells, nil, nil, nil, nil))
	require.NoError(t, err)
	require.Len(t, res, 1)

	// Altitude window entirely above the operational intent excludes it.
	var lo float32 = 200
	res, err = r.SearchOperationalIntents(ctx, volume4D(cells, nil, nil, &lo, nil))
	require.NoError(t, err)
	require.Empty(t, res)

	// Missing footprint is a bad request.
	_, err = r.SearchOperationalIntents(ctx, &dssmodels.Volume4D{})
	require.Error(t, err)
}

func TestGetDependentOperationalIntents(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)
	_, err := r.UpsertOperationalIntent(ctx, sampleOperationalIntent())
	require.NoError(t, err)

	deps, err := r.GetDependentOperationalIntents(ctx, subscriptionId)
	require.NoError(t, err)
	require.Equal(t, []dssmodels.ID{operationalIntentId}, deps)

	deps, err = r.GetDependentOperationalIntents(ctx, "other")
	require.NoError(t, err)
	require.Nil(t, deps)
}

var (
	oi1ID = dssmodels.ID("00000185-e36d-40be-8d38-beca6ca30000")
	oi2ID = dssmodels.ID("00000185-e36d-40be-8d38-beca6ca30001")
	oi3ID = dssmodels.ID("00000185-e36d-40be-8d38-beca6ca30003")

	start1 = time.Date(2024, time.August, 14, 15, 48, 36, 0, time.UTC)
	end1   = start1.Add(time.Hour)
	start2 = time.Date(2024, time.September, 15, 15, 48, 36, 0, time.UTC)
	end2   = start2.Add(time.Hour)
	start3 = time.Date(2024, time.September, 16, 15, 48, 36, 0, time.UTC)
	end3   = start3.Add(time.Hour)
)

var (
	oi1 = &scdmodels.OperationalIntent{
		ID:             oi1ID,
		Manager:        "unittest",
		Version:        1,
		State:          scdmodels.OperationalIntentStateAccepted,
		StartTime:      &start1,
		EndTime:        &end1,
		USSBaseURL:     "https://dummy.uss",
		SubscriptionID: &sub1ID,
		AltitudeLower:  &altLow,
		AltitudeUpper:  &altHigh,
		Cells:          cells,
	}
	oi2 = &scdmodels.OperationalIntent{
		ID:             oi2ID,
		Manager:        "unittest",
		Version:        1,
		State:          scdmodels.OperationalIntentStateAccepted,
		StartTime:      &start2,
		EndTime:        &end2,
		USSBaseURL:     "https://dummy.uss",
		SubscriptionID: &sub2ID,
		AltitudeLower:  &altLow,
		AltitudeUpper:  &altHigh,
		Cells:          cells,
	}
	oi3 = &scdmodels.OperationalIntent{
		ID:             oi3ID,
		Manager:        "unittest",
		Version:        1,
		State:          scdmodels.OperationalIntentStateAccepted,
		StartTime:      &start3,
		EndTime:        &end3,
		USSBaseURL:     "https://dummy.uss",
		SubscriptionID: &sub3ID,
		AltitudeLower:  &altLow,
		AltitudeUpper:  &altHigh,
		Cells:          cells,
	}
)

func TestListExpiredOperationalIntents(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)

	_, err := r.UpsertSubscription(ctx, sub1)
	require.NoError(t, err)
	_, err = r.UpsertOperationalIntent(ctx, oi1)
	require.NoError(t, err)

	_, err = r.UpsertSubscription(ctx, sub2)
	require.NoError(t, err)
	_, err = r.UpsertOperationalIntent(ctx, oi2)
	require.NoError(t, err)

	_, err = r.UpsertSubscription(ctx, sub3)
	require.NoError(t, err)
	_, err = r.UpsertOperationalIntent(ctx, oi3)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		timeRef time.Time
		ttl     time.Duration
		expired []dssmodels.ID
	}{{
		name:    "none expired, one in close past",
		timeRef: time.Date(2024, time.August, 25, 15, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 24 * 30,
		expired: []dssmodels.ID{},
	}, {
		name:    "one recently expired, one current, one in future",
		timeRef: time.Date(2024, time.September, 15, 16, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 24 * 30,
		expired: []dssmodels.ID{oi1ID},
	}, {
		name:    "two expired, one in future",
		timeRef: time.Date(2024, time.September, 16, 16, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 2,
		expired: []dssmodels.ID{oi1ID, oi2ID},
	}, {
		name:    "all expired",
		timeRef: time.Date(2024, time.December, 15, 15, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 24 * 30,
		expired: []dssmodels.ID{oi1ID, oi2ID, oi3ID},
	}}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			threshold := testCase.timeRef.Add(-testCase.ttl)
			expired, err := r.ListExpiredOperationalIntents(ctx, threshold)
			require.NoError(t, err)

			expiredIDs := make([]dssmodels.ID, 0, len(expired))
			for _, expiredOi := range expired {
				expiredIDs = append(expiredIDs, expiredOi.ID)
			}
			require.ElementsMatch(t, expiredIDs, testCase.expired)
		})
	}
}
