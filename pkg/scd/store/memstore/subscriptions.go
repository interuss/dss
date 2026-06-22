package memstore

import (
	"context"
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
		StartTime:                   cloneTime(rec.StartTime),
		EndTime:                     cloneTime(rec.EndTime),
		USSBaseURL:                  rec.USSBaseURL,
		NotifyForOperationalIntents: rec.NotifyForOperationalIntents,
		NotifyForConstraints:        rec.NotifyForConstraints,
		ImplicitSubscription:        rec.ImplicitSubscription,
		Cells:                       cloneCells(rec.Cells),
	}
}

// SearchSubscriptions implements scd.repos.Subscription.SearchSubscriptions.
func (r *repo) SearchSubscriptions(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	cells, err := v4d.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not calculate spatial covering")
	}
	if len(cells) == 0 {
		return nil, nil
	}

	want := cellSet(cells)
	var out []*scdmodels.Subscription
	for _, rec := range r.state.Subscriptions {
		if !overlaps(rec.Cells, want) {
			continue
		}
		// COALESCE(starts_at <= $3, true) with $3 = v4d.EndTime
		if rec.StartTime != nil && v4d.EndTime != nil && rec.StartTime.After(*v4d.EndTime) {
			continue
		}
		// COALESCE(ends_at >= $2, true) with $2 = v4d.StartTime
		if rec.EndTime != nil && v4d.StartTime != nil && rec.EndTime.Before(*v4d.StartTime) {
			continue
		}
		out = append(out, rec.toModel())
		if len(out) >= dssmodels.MaxResultLimit { // mirror SQL "LIMIT MaxResultLimit"
			break
		}
	}
	return out, nil
}

// GetSubscription implements scd.repos.Subscription.GetSubscription.
func (r *repo) GetSubscription(_ context.Context, id dssmodels.ID) (*scdmodels.Subscription, error) {
	rec, ok := r.state.Subscriptions[id]
	if !ok {
		return nil, nil
	}
	return rec.toModel(), nil
}

// UpsertSubscription implements scd.repos.Subscription.UpsertSubscription.
func (r *repo) UpsertSubscription(ctx context.Context, s *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	rec := &subscriptionRecord{
		ID:                          s.ID,
		Manager:                     s.Manager,
		NotificationIndex:           s.NotificationIndex,
		USSBaseURL:                  s.USSBaseURL,
		NotifyForOperationalIntents: s.NotifyForOperationalIntents,
		NotifyForConstraints:        s.NotifyForConstraints,
		ImplicitSubscription:        s.ImplicitSubscription,
		StartTime:                   cloneTime(s.StartTime),
		EndTime:                     cloneTime(s.EndTime),
		Cells:                       cloneCells(s.Cells),
		UpdatedAt:                   timestamp.NowFromContext(ctx),
	}
	r.state.Subscriptions[s.ID] = rec
	return rec.toModel(), nil
}

// DeleteSubscription implements scd.repos.Subscription.DeleteSubscription.
func (r *repo) DeleteSubscription(_ context.Context, id dssmodels.ID) error {
	if _, ok := r.state.Subscriptions[id]; !ok {
		return stacktrace.NewError("Attempted to delete non-existent Subscription")
	}
	delete(r.state.Subscriptions, id)
	return nil
}

// IncrementNotificationIndicesForOperationalIntents implements
// scd.repos.Subscription.IncrementNotificationIndicesForOperationalIntents.
func (r *repo) IncrementNotificationIndicesForOperationalIntents(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	cells, err := v4d.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not calculate spatial covering")
	}
	if len(cells) == 0 {
		return nil, nil
	}

	want := cellSet(cells)
	var out []*scdmodels.Subscription
	for _, rec := range r.state.Subscriptions {
		if !overlaps(rec.Cells, want) {
			continue
		}
		if !rec.NotifyForOperationalIntents {
			continue
		}
		// COALESCE(starts_at <= $3, true) with $3 = v4d.EndTime
		if rec.StartTime != nil && v4d.EndTime != nil && rec.StartTime.After(*v4d.EndTime) {
			continue
		}
		// COALESCE(ends_at >= $2, true) with $2 = v4d.StartTime
		if rec.EndTime != nil && v4d.StartTime != nil && rec.EndTime.Before(*v4d.StartTime) {
			continue
		}
		rec.NotificationIndex++
		out = append(out, rec.toModel())
	}
	return out, nil
}

// IncrementNotificationIndicesForConstraints implements
// scd.repos.Subscription.IncrementNotificationIndicesForConstraints.
func (r *repo) IncrementNotificationIndicesForConstraints(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error) {
	cells, err := v4d.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not calculate spatial covering")
	}
	if len(cells) == 0 {
		return nil, nil
	}

	want := cellSet(cells)
	var out []*scdmodels.Subscription
	for _, rec := range r.state.Subscriptions {
		if !overlaps(rec.Cells, want) {
			continue
		}
		if !rec.NotifyForConstraints {
			continue
		}
		// COALESCE(starts_at <= $3, true) with $3 = v4d.EndTime
		if rec.StartTime != nil && v4d.EndTime != nil && rec.StartTime.After(*v4d.EndTime) {
			continue
		}
		// COALESCE(ends_at >= $2, true) with $2 = v4d.StartTime
		if rec.EndTime != nil && v4d.StartTime != nil && rec.EndTime.Before(*v4d.StartTime) {
			continue
		}
		rec.NotificationIndex++
		out = append(out, rec.toModel())
	}
	return out, nil
}

// LockSubscriptionsOnCells implements scd.repos.Subscription.LockSubscriptionsOnCells.
func (r *repo) LockSubscriptionsOnCells(_ context.Context, _ s2.CellUnion, _ []dssmodels.ID, _ *time.Time, _ *time.Time) error {
	// For the memory store, that a no-op
	return nil
}

// ListExpiredSubscriptions implements scd.repos.Subscription.ListExpiredSubscriptions.
func (r *repo) ListExpiredSubscriptions(_ context.Context, threshold time.Time) ([]*scdmodels.Subscription, error) {
	var out []*scdmodels.Subscription
	for _, rec := range r.state.Subscriptions {
		// (ends_at IS NOT NULL AND ends_at <= threshold) OR (ends_at IS NULL AND updated_at <= threshold)
		var expired bool
		if rec.EndTime != nil {
			expired = !rec.EndTime.After(threshold)
		} else {
			expired = !rec.UpdatedAt.After(threshold)
		}
		if !expired {
			continue
		}
		out = append(out, rec.toModel())
		if len(out) >= dssmodels.MaxResultLimit { // mirror SQL "LIMIT MaxResultLimit"
			break
		}
	}
	return out, nil
}

func (r *repo) CountSubscriptions(_ context.Context) (int64, error) {
	return int64(len(r.state.Subscriptions)), nil
}
