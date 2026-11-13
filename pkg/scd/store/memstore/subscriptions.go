package memstore

import (
	"context"
	"iter"
	"maps"
	"slices"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

func (rec *subscriptionRecord) toModel() *scdmodels.Subscription {
	return &scdmodels.Subscription{
		ID:                          rec.ID,
		Version:                     scdmodels.NewOVNFromTime(rec.UpdatedAt, rec.ID.String()),
		NotificationIndex:           rec.NotificationIndex,
		Manager:                     rec.Manager,
		USSBaseURL:                  rec.USSBaseURL,
		NotifyForOperationalIntents: rec.NotifyForOperationalIntents,
		NotifyForConstraints:        rec.NotifyForConstraints,
		ImplicitSubscription:        rec.ImplicitSubscription,
		CellsVolume4D: &dssmodels.CellsVolume4D{
			StartTime: new(rec.StartTime),
			EndTime:   new(rec.EndTime),
			Cells:     slices.Clone(rec.Cells),
		},
	}
}

// subscriptionsInCellsVolume yields the subscriptions intersecting cellsVolume.
func (r *repo) subscriptionsInCellsVolume(cellsVolume *dssmodels.CellsVolume4D) iter.Seq[*subscriptionRecord] {
	want := cellSet(cellsVolume.Cells)

	return func(yield func(*subscriptionRecord) bool) {
		for _, rec := range r.state.Subscriptions {
			if !overlaps(rec.Cells, want) {
				continue
			}
			if !overlapsTime(rec.StartTime, rec.EndTime, cellsVolume) {
				continue
			}
			if !yield(rec) {
				return
			}
		}
	}
}

func (r *repo) SearchSubscriptions(_ context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Subscription, error) {
	subscriptions := r.subscriptionsInCellsVolume(cellsVolume)

	var out []*scdmodels.Subscription
	for rec := range subscriptions {
		out = append(out, rec.toModel())

		if len(out) >= dssmodels.MaxResultLimit { // mirror SQL "LIMIT MaxResultLimit"
			break
		}
	}
	return out, nil
}

func (r *repo) GetSubscription(_ context.Context, id dssmodels.ID) (*scdmodels.Subscription, error) {
	rec, ok := r.state.Subscriptions[id]
	if !ok {
		return nil, nil
	}
	return rec.toModel(), nil
}

func (r *repo) UpsertSubscription(ctx context.Context, s *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	if err := requireExtentTimes(s.StartTime, s.EndTime); err != nil {
		return nil, err
	}

	now := timestamp.MustFromContext(ctx)

	rec := &subscriptionRecord{
		ID:                          s.ID,
		Manager:                     s.Manager,
		NotificationIndex:           s.NotificationIndex,
		USSBaseURL:                  s.USSBaseURL,
		NotifyForOperationalIntents: s.NotifyForOperationalIntents,
		NotifyForConstraints:        s.NotifyForConstraints,
		ImplicitSubscription:        s.ImplicitSubscription,
		StartTime:                   *s.StartTime,
		EndTime:                     *s.EndTime,
		Cells:                       slices.Clone(s.Cells),
		UpdatedAt:                   now,
	}
	r.state.Subscriptions[s.ID] = rec
	return rec.toModel(), nil
}

func (r *repo) DeleteSubscription(_ context.Context, id dssmodels.ID) error {
	if _, ok := r.state.Subscriptions[id]; !ok {
		return stacktrace.NewError("Attempted to delete non-existent Subscription")
	}
	delete(r.state.Subscriptions, id)
	// Delete any operational intents referencing the subscription.
	// Mirrors the behavior of the sql store, which has a foreign key constraint with cascade delete.
	r.deleteOperationalIntentsOfSubscriptions([]dssmodels.ID{id})
	return nil
}

// deleteOperationalIntentsOfSubscriptions deletes the operational intents referencing the given subscriptions
func (r *repo) deleteOperationalIntentsOfSubscriptions(ids []dssmodels.ID) {
	deleted := make(map[dssmodels.ID]struct{}, len(ids))
	for _, id := range ids {
		deleted[id] = struct{}{}
	}
	maps.DeleteFunc(r.state.OperationalIntents, func(_ dssmodels.ID, rec *operationalIntentRecord) bool {
		if rec.SubscriptionID == nil {
			return false
		}
		_, ok := deleted[*rec.SubscriptionID]
		return ok
	})
}

func (r *repo) IncrementNotificationIndicesForOperationalIntents(_ context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Subscription, error) {
	return r.incrementNotificationIndices(cellsVolume, func(rec *subscriptionRecord) bool { return rec.NotifyForOperationalIntents })
}

func (r *repo) IncrementNotificationIndicesForConstraints(_ context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Subscription, error) {
	return r.incrementNotificationIndices(cellsVolume, func(rec *subscriptionRecord) bool { return rec.NotifyForConstraints })
}

// incrementNotificationIndices increments the notification index of each subscription
// intersecting cellsVolume and asking for the notifications selected by notified.
func (r *repo) incrementNotificationIndices(cellsVolume *dssmodels.CellsVolume4D, notified func(*subscriptionRecord) bool) ([]*scdmodels.Subscription, error) {
	subscriptions := r.subscriptionsInCellsVolume(cellsVolume)

	var out []*scdmodels.Subscription
	for rec := range subscriptions {
		if !notified(rec) {
			continue
		}
		rec.NotificationIndex++
		out = append(out, rec.toModel())
	}
	return out, nil
}

func (r *repo) LockSubscriptionsOnCells(_ context.Context, _ s2.CellUnion, _ []dssmodels.ID, _ *time.Time, _ *time.Time) error {
	// For the memory store, that a no-op
	return nil
}

func (r *repo) ListExpiredSubscriptions(_ context.Context, threshold time.Time) ([]dssmodels.ID, error) {
	return listExpired(r.state.Subscriptions, threshold), nil
}

func (r *repo) DeleteExpiredSubscriptions(_ context.Context, threshold time.Time, limit int) ([]dssmodels.ID, error) {
	ids := deleteExpired(r.state.Subscriptions, threshold, limit)
	// Delete any operational intents referencing the subscription.
	// Mirrors the behavior of the sql store, which has a foreign key constraint with cascade delete.
	r.deleteOperationalIntentsOfSubscriptions(ids)
	return ids, nil
}

func (r *repo) CountSubscriptions(_ context.Context) (int64, error) {
	return int64(len(r.state.Subscriptions)), nil
}
