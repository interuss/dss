package cockroach

import (
	"context"
	"testing"
	"time"

	"github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/stretchr/testify/require"
)

var (
	sub1ID = models.ID("189ec22f-5e61-418a-940b-36de2d201fd5")
	sub2ID = models.ID("78f98cc5-94f3-4c04-8da9-a8398feba3f3")
	sub3ID = models.ID("9f0d4575-b275-4a4c-a261-e1e04d324565")
)

var (
	sub1 = &scdmodels.Subscription{
		ID:                          sub1ID,
		NotificationIndex:           1,
		Manager:                     "unittest",
		StartTime:                   &start1,
		EndTime:                     &end1,
		USSBaseURL:                  "https://dummy.uss",
		NotifyForOperationalIntents: true,
		NotifyForConstraints:        false,
		ImplicitSubscription:        true,
		Cells:                       cells,
	}
	sub2 = &scdmodels.Subscription{
		ID:                          sub2ID,
		NotificationIndex:           1,
		Manager:                     "unittest",
		StartTime:                   &start2,
		EndTime:                     &end2,
		USSBaseURL:                  "https://dummy.uss",
		NotifyForOperationalIntents: true,
		NotifyForConstraints:        false,
		ImplicitSubscription:        true,
		Cells:                       cells,
	}
	sub3 = &scdmodels.Subscription{
		ID:                          sub3ID,
		NotificationIndex:           1,
		Manager:                     "unittest",
		StartTime:                   &start3,
		EndTime:                     &end3,
		USSBaseURL:                  "https://dummy.uss",
		NotifyForOperationalIntents: true,
		NotifyForConstraints:        false,
		ImplicitSubscription:        true,
		Cells:                       cells,
	}
)

func TestListExpiredSubscriptions(t *testing.T) {
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

	_, err = r.UpsertSubscription(ctx, sub2)
	require.NoError(t, err)

	_, err = r.UpsertSubscription(ctx, sub3)
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
		expired: []models.ID{sub1ID},
	}, {
		name:    "two expired, one in future",
		timeRef: time.Date(2024, time.September, 16, 16, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 2,
		expired: []models.ID{sub1ID, sub2ID},
	}, {
		name:    "all expired",
		timeRef: time.Date(2024, time.December, 15, 15, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 24 * 30,
		expired: []models.ID{sub1ID, sub2ID, sub3ID},
	}}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			threshold := testCase.timeRef.Add(-testCase.ttl)
			expired, err := r.ListExpiredSubscriptions(ctx, threshold)
			require.NoError(t, err)

			expiredIDs := make([]models.ID, 0, len(expired))
			for _, expiredSub := range expired {
				expiredIDs = append(expiredIDs, expiredSub.ID)
			}
			require.ElementsMatch(t, expiredIDs, testCase.expired)
		})
	}
}
