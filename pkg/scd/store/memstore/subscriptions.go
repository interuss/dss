package memstore

import (
	"context"
	"iter"
	"slices"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/memstore/utils"
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
		StartTime:                   utils.ClonePtr(rec.StartTime),
		EndTime:                     utils.ClonePtr(rec.EndTime),
		USSBaseURL:                  rec.USSBaseURL,
		NotifyForOperationalIntents: rec.NotifyForOperationalIntents,
		NotifyForConstraints:        rec.NotifyForConstraints,
		ImplicitSubscription:        rec.ImplicitSubscription,
		Cells:                       slices.Clone(rec.Cells),
	}
}

// subscriptionsInVolume4D yields the subscriptions intersecting v4d.
func (r *repo) subscriptionsInVolume4D(v4d *dssmodels.Volume4D) (iter.Seq[*subscriptionRecord], error) {
	want, err := coveringSet(v4d)
	if err != nil {
		return nil, err
	}

	return func(yield func(*subscriptionRecord) bool) {
		for _, rec := range r.state.Subscriptions {
			if !overlaps(rec.Cells, want) {
				continue
			}
			if !overlapsTime(rec.StartTime, rec.EndTime, v4d) {
				continue
			}
			if !yield(rec) {
				return
			}
		}
	}, nil
}

func (r *repo) SearchSubscriptions(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	subscriptions, err := r.subscriptionsInVolume4D(v4d)
	if err != nil {
		return nil, err
	}

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
	now := timestamp.MustFromContext(ctx)

	rec := &subscriptionRecord{
		ID:                          s.ID,
		Manager:                     s.Manager,
		NotificationIndex:           s.NotificationIndex,
		USSBaseURL:                  s.USSBaseURL,
		NotifyForOperationalIntents: s.NotifyForOperationalIntents,
		NotifyForConstraints:        s.NotifyForConstraints,
		ImplicitSubscription:        s.ImplicitSubscription,
		StartTime:                   utils.ClonePtr(s.StartTime),
		EndTime:                     utils.ClonePtr(s.EndTime),
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
	return nil
}

func (r *repo) IncrementNotificationIndicesForOperationalIntents(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	return r.incrementNotificationIndices(v4d, func(rec *subscriptionRecord) bool { return rec.NotifyForOperationalIntents })
}

func (r *repo) IncrementNotificationIndicesForConstraints(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	return r.incrementNotificationIndices(v4d, func(rec *subscriptionRecord) bool { return rec.NotifyForConstraints })
}

// incrementNotificationIndices increments the notification index of each subscription
// intersecting v4d and asking for the notifications selected by notified.
func (r *repo) incrementNotificationIndices(v4d *dssmodels.Volume4D, notified func(*subscriptionRecord) bool) ([]*scdmodels.Subscription, error) {
	subscriptions, err := r.subscriptionsInVolume4D(v4d)
	if err != nil {
		return nil, err
	}

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

func (r *repo) ListExpiredSubscriptions(_ context.Context, threshold time.Time) ([]*scdmodels.Subscription, error) {
	var out []*scdmodels.Subscription
	for _, rec := range listExpired(r.state.Subscriptions, threshold, dssmodels.MaxResultLimit) {
		out = append(out, rec.toModel())
	}
	return out, nil
}

func (r *repo) CountSubscriptions(_ context.Context) (int64, error) {
	return int64(len(r.state.Subscriptions)), nil
}
