package operations

import (
	"context"
	"time"

	"testing"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/require"
)

type fakeISARepo struct {
	isas map[dssmodels.ID]*ridmodels.IdentificationServiceArea
	subs map[dssmodels.ID]*ridmodels.Subscription
}

func newFakeISARepo() *fakeISARepo {
	return &fakeISARepo{
		isas: make(map[dssmodels.ID]*ridmodels.IdentificationServiceArea),
		subs: make(map[dssmodels.ID]*ridmodels.Subscription),
	}
}

func (r *fakeISARepo) GetISA(_ context.Context, id dssmodels.ID, _ bool) (*ridmodels.IdentificationServiceArea, error) {
	if isa, ok := r.isas[id]; ok {
		return isa, nil
	}
	return nil, nil
}

func (r *fakeISARepo) DeleteISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	if existing, ok := r.isas[isa.ID]; ok {
		delete(r.isas, isa.ID)
		return existing, nil
	}
	return nil, nil
}

func (r *fakeISARepo) InsertISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	storedCopy := *isa
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	r.isas[isa.ID] = &storedCopy
	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (r *fakeISARepo) UpdateISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	storedCopy := *isa
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	r.isas[isa.ID] = &storedCopy
	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (r *fakeISARepo) SearchISAs(_ context.Context, _ s2.CellUnion, _ *time.Time, _ *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	panic("not implemented")
}

func (r *fakeISARepo) ListExpiredISAs(_ context.Context, _ string, _ time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	panic("not implemented")
}

func (r *fakeISARepo) CountISAs(_ context.Context) (int64, error) {
	panic("not implemented")
}

func (r *fakeISARepo) GetSubscription(_ context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	if sub, ok := r.subs[id]; ok {
		return sub, nil
	}
	return nil, nil
}

func (r *fakeISARepo) DeleteSubscription(_ context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	if sub, ok := r.subs[s.ID]; ok {
		delete(r.subs, s.ID)
		return sub, nil
	}
	return nil, nil
}

func (r *fakeISARepo) InsertSubscription(_ context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	storedCopy := *s
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	r.subs[s.ID] = &storedCopy
	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (r *fakeISARepo) UpdateSubscription(_ context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	storedCopy := *s
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	r.subs[s.ID] = &storedCopy
	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (r *fakeISARepo) SearchSubscriptions(_ context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	var subs []*ridmodels.Subscription
	for _, s := range r.subs {
		for _, c1 := range s.Cells {
			for _, c2 := range cells {
				if c1 == c2 {
					subs = append(subs, s)
					break
				}
			}
		}
	}
	return subs, nil
}

func (r *fakeISARepo) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	res, err := r.SearchSubscriptions(ctx, cells)
	if err != nil {
		return nil, err
	}
	var subs []*ridmodels.Subscription
	for _, s := range res {
		if s.Owner == owner {
			subs = append(subs, s)
		}
	}
	return subs, nil
}

func (r *fakeISARepo) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	subs, err := r.SearchSubscriptions(ctx, cells)
	if err != nil {
		return nil, err
	}
	for i := range subs {
		subs[i].NotificationIndex++
	}
	return subs, nil
}

func (r *fakeISARepo) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	maxValue := 0
	subs, err := r.SearchSubscriptionsByOwner(ctx, cells, owner)
	if err != nil {
		return 0, err
	}

	cellMap := make(map[s2.CellID]int)
	for _, s := range subs {
		for _, cid := range s.Cells {
			cellMap[cid]++
			if cellMap[cid] > maxValue {
				maxValue = cellMap[cid]
			}
		}
	}
	return maxValue, nil
}

func (r *fakeISARepo) ListExpiredSubscriptions(_ context.Context, _ string, _ time.Time) ([]*ridmodels.Subscription, error) {
	panic("not implemented")
}

func (r *fakeISARepo) CountSubscriptions(_ context.Context) (int64, error) {
	panic("not implemented")
}

func TestDeleteISA(t *testing.T) {
	var (
		ctx  = newTestContext()
		repo = newFakeISARepo()
	)

	insertedSubscriptions := []*ridmodels.Subscription{}
	for _, r := range subscriptionsPool {
		subscriptionCopy := *r.input
		s1, err := InsertSubscription(ctx, repo, &subscriptionCopy)
		require.NoError(t, err)
		require.NotNil(t, s1)
		require.Equal(t, 42, s1.NotificationIndex)
		insertedSubscriptions = append(insertedSubscriptions, s1)
	}
	serviceArea := &ridmodels.IdentificationServiceArea{
		ID:        dssmodels.ID(uuid.New().String()),
		Owner:     dssmodels.Owner(uuid.New().String()),
		URL:       "https://no/place/like/home/for/flights",
		StartTime: &startTime,
		EndTime:   &endTime,
		Cells: s2.CellUnion{
			s2.CellID(12494535935418957824),
		},
	}

	// Insert the ISA
	serviceAreaCopy := *serviceArea
	subscriptionsOut, err := repo.UpdateNotificationIdxsInCells(ctx, serviceAreaCopy.Cells)
	require.NoError(t, err)
	isa, err := repo.InsertISA(ctx, &serviceAreaCopy)
	require.NoError(t, err)
	require.NotNil(t, isa)
	require.Len(t, subscriptionsOut, len(insertedSubscriptions))

	for i := range insertedSubscriptions {
		require.Equal(t, 43, subscriptionsOut[i].NotificationIndex)
	}

	// Can't delete with different owner.
	_, err = deleteISA(ctx, repo, isa.ID, "bad-owner", isa.Version)
	require.Error(t, err)
	require.Equal(t, dsserr.PermissionDenied, stacktrace.GetCode(err))

	// Delete the ISA.
	// Ensure a fresh Get, then delete still updates the subscription indexes.
	isa, err = repo.GetISA(ctx, isa.ID, false)
	require.NoError(t, err)

	result, err := deleteISA(ctx, repo, isa.ID, isa.Owner, isa.Version)
	require.NoError(t, err)
	require.Equal(t, isa, result.ISA)
	require.NotNil(t, result.Subscriptions)
	require.Len(t, result.Subscriptions, len(subscriptionsPool))
	for i, s := range subscriptionsPool {
		require.Equal(t, s.input.URL, result.Subscriptions[i].URL)
	}

	for i := range insertedSubscriptions {
		require.Equal(t, 44, result.Subscriptions[i].NotificationIndex)
	}
}
