package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/stretchr/testify/require"
)

var (
	oi1ID = models.ID("00000185-e36d-40be-8d38-beca6ca30000")
	oi2ID = models.ID("00000185-e36d-40be-8d38-beca6ca30001")
	oi3ID = models.ID("00000185-e36d-40be-8d38-beca6ca30003")

	cells = s2.CellUnion{
		s2.CellID(int64(8768904281496485888)),
		s2.CellID(int64(8768904178417270784)),
	}

	start1 = time.Date(2024, time.August, 14, 15, 48, 36, 0, time.UTC)
	end1   = start1.Add(time.Hour)
	start2 = time.Date(2024, time.September, 15, 15, 48, 36, 0, time.UTC)
	end2   = start2.Add(time.Hour)
	start3 = time.Date(2024, time.September, 16, 15, 48, 36, 0, time.UTC)
	end3   = start3.Add(time.Hour)

	altLow, altHigh float32 = 84, 169
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
	var (
		ctx                  = context.Background()
		store, tearDownStore = setUpStore(ctx, t)
	)
	require.NotNil(t, store)
	defer tearDownStore()

	r, err := store.Interact(ctx)
	require.NoError(t, err)

	_, err = r.UpsertSubscription(ctx, sub1)
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
		expired []models.ID
	}{{
		name:    "none expired, one in close past",
		timeRef: time.Date(2024, time.August, 25, 15, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 24 * 30,
		expired: []models.ID{},
	}, {
		name:    "one recently expired, one current, one in future",
		timeRef: time.Date(2024, time.September, 15, 16, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 24 * 30,
		expired: []models.ID{oi1ID},
	}, {
		name:    "two expired, one in future",
		timeRef: time.Date(2024, time.September, 16, 16, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 2,
		expired: []models.ID{oi1ID, oi2ID},
	}, {
		name:    "all expired",
		timeRef: time.Date(2024, time.December, 15, 15, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 24 * 30,
		expired: []models.ID{oi1ID, oi2ID, oi3ID},
	}}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			threshold := testCase.timeRef.Add(-testCase.ttl)
			expired, err := r.ListExpiredOperationalIntents(ctx, threshold)
			require.NoError(t, err)

			expiredIDs := make([]models.ID, 0, len(expired))
			for _, expiredOi := range expired {
				expiredIDs = append(expiredIDs, expiredOi.ID)
			}
			require.ElementsMatch(t, expiredIDs, testCase.expired)
		})
	}
}
