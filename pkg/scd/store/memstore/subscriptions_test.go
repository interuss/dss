package memstore

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionUpsertGetDelete(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)

	got, err := r.UpsertSubscription(ctx, sampleSubscription())
	require.NoError(t, err)
	require.Equal(t, subscriptionId, got.ID)
	require.Equal(t, 1, got.NotificationIndex)
	require.NotEmpty(t, got.Version)

	fetched, err := r.GetSubscription(ctx, subscriptionId)
	require.NoError(t, err)
	require.Equal(t, got.Version, fetched.Version)
	require.True(t, fetched.NotifyForOperationalIntents)

	count, err := r.CountSubscriptions(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	require.NoError(t, r.DeleteSubscription(ctx, subscriptionId))
	gone, err := r.GetSubscription(ctx, subscriptionId)
	require.NoError(t, err)
	require.Nil(t, gone)
}

func TestSubscriptionGetMissingReturnsNil(t *testing.T) {
	r := setUpStore(t)
	got, err := r.GetSubscription(writeCtx(), subscriptionId)
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestSubscriptionDeleteMissingErrors(t *testing.T) {
	r := setUpStore(t)
	require.Error(t, r.DeleteSubscription(writeCtx(), subscriptionId))
}

func TestSearchSubscriptions(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)
	_, err := r.UpsertSubscription(ctx, sampleSubscription())
	require.NoError(t, err)

	res, err := r.SearchSubscriptions(ctx, volume4D(cells, nil, nil, nil, nil))
	require.NoError(t, err)
	require.Len(t, res, 1)

	// No covering cells returns nil.
	res, err = r.SearchSubscriptions(ctx, volume4D(s2.CellUnion{}, nil, nil, nil, nil))
	require.NoError(t, err)
	require.Nil(t, res)
}

func TestIncrementNotificationIndicesForOperationalIntents(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)

	// notify_for_operations = true.
	opSub := sampleSubscription()
	_, err := r.UpsertSubscription(ctx, opSub)
	require.NoError(t, err)

	// A second subscription that only wants constraint notifications must be skipped.
	conSub := sampleSubscription()
	conSub.ID = "00000185-e36d-40be-8d38-beca6ca31aaa"
	conSub.NotifyForOperationalIntents = false
	conSub.NotifyForConstraints = true
	_, err = r.UpsertSubscription(ctx, conSub)
	require.NoError(t, err)

	got, err := r.IncrementNotificationIndicesForOperationalIntents(ctx, volume4D(cells, nil, nil, nil, nil))
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, opSub.ID, got[0].ID)
	require.Equal(t, opSub.NotificationIndex+1, got[0].NotificationIndex)

	// The bump is persisted.
	fetched, err := r.GetSubscription(ctx, opSub.ID)
	require.NoError(t, err)
	require.Equal(t, opSub.NotificationIndex+1, fetched.NotificationIndex)

	// The constraint-only subscription was untouched.
	other, err := r.GetSubscription(ctx, conSub.ID)
	require.NoError(t, err)
	require.Equal(t, conSub.NotificationIndex, other.NotificationIndex)

	// No covering cells returns nil.
	got, err = r.IncrementNotificationIndicesForOperationalIntents(ctx, volume4D(s2.CellUnion{}, nil, nil, nil, nil))
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestIncrementNotificationIndicesForConstraints(t *testing.T) {
	ctx := writeCtx()
	r := setUpStore(t)

	// notify_for_constraints = true (sample sets both notify flags).
	conSub := sampleSubscription()
	_, err := r.UpsertSubscription(ctx, conSub)
	require.NoError(t, err)

	// A subscription that does not want constraint notifications must be skipped.
	opSub := sampleSubscription()
	opSub.ID = "00000185-e36d-40be-8d38-beca6ca31bbb"
	opSub.NotifyForConstraints = false
	_, err = r.UpsertSubscription(ctx, opSub)
	require.NoError(t, err)

	got, err := r.IncrementNotificationIndicesForConstraints(ctx, volume4D(cells, nil, nil, nil, nil))
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, conSub.ID, got[0].ID)
	require.Equal(t, conSub.NotificationIndex+1, got[0].NotificationIndex)

	other, err := r.GetSubscription(ctx, opSub.ID)
	require.NoError(t, err)
	require.Equal(t, opSub.NotificationIndex, other.NotificationIndex)
}

func TestLockSubscriptionsOnCellsNoop(t *testing.T) {
	r := setUpStore(t)
	require.NoError(t, r.LockSubscriptionsOnCells(writeCtx(), cells, []dssmodels.ID{subscriptionId}, nil, nil))
}

var (
	sub1ID = dssmodels.ID("189ec22f-5e61-418a-940b-36de2d201fd5")
	sub2ID = dssmodels.ID("78f98cc5-94f3-4c04-8da9-a8398feba3f3")
	sub3ID = dssmodels.ID("9f0d4575-b275-4a4c-a261-e1e04d324565")
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
	ctx := writeCtx()
	r := setUpStore(t)

	_, err := r.UpsertSubscription(ctx, sub1)
	require.NoError(t, err)

	_, err = r.UpsertSubscription(ctx, sub2)
	require.NoError(t, err)

	_, err = r.UpsertSubscription(ctx, sub3)
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
		expired: []dssmodels.ID{sub1ID},
	}, {
		name:    "two expired, one in future",
		timeRef: time.Date(2024, time.September, 16, 16, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 2,
		expired: []dssmodels.ID{sub1ID, sub2ID},
	}, {
		name:    "all expired",
		timeRef: time.Date(2024, time.December, 15, 15, 0, 0, 0, time.UTC),
		ttl:     time.Hour * 24 * 30,
		expired: []dssmodels.ID{sub1ID, sub2ID, sub3ID},
	}}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			threshold := testCase.timeRef.Add(-testCase.ttl)
			expired, err := r.ListExpiredSubscriptions(ctx, threshold)
			require.NoError(t, err)

			expiredIDs := make([]dssmodels.ID, 0, len(expired))
			for _, expiredSub := range expired {
				expiredIDs = append(expiredIDs, expiredSub.ID)
			}
			require.ElementsMatch(t, expiredIDs, testCase.expired)
		})
	}
}
