package memstore

// Note: as of now, doesn't implement timeBasedNotificationIndex settings, as it doesn't improve performance
// and was done mostly for SQL store improvements.

import (
	"context"
	"iter"
	"maps"
	"slices"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/memstore/utils"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

func subRecordFromModel(s *ridmodels.Subscription, updatedAt time.Time) *subscriptionRecord {
	return &subscriptionRecord{
		ID:                s.ID,
		URL:               s.URL,
		NotificationIndex: s.NotificationIndex,
		Owner:             s.Owner,
		Cells:             slices.Clone(s.Cells),
		StartTime:         utils.ClonePtr(s.StartTime),
		EndTime:           utils.ClonePtr(s.EndTime),
		AltitudeHi:        utils.ClonePtr(s.AltitudeHi), // TODO: As noted during review, altitudes seems unused.
		AltitudeLo:        utils.ClonePtr(s.AltitudeLo),
		Writer:            s.Writer,
		UpdatedAt:         updatedAt,
	}
}

func (rec *subscriptionRecord) toModel() *ridmodels.Subscription {
	return &ridmodels.Subscription{
		ID:                rec.ID,
		URL:               rec.URL,
		NotificationIndex: rec.NotificationIndex,
		Owner:             rec.Owner,
		Cells:             slices.Clone(rec.Cells),
		StartTime:         utils.ClonePtr(rec.StartTime),
		EndTime:           utils.ClonePtr(rec.EndTime),
		Version:           dssmodels.VersionFromTime(rec.UpdatedAt),
		AltitudeHi:        utils.ClonePtr(rec.AltitudeHi),
		AltitudeLo:        utils.ClonePtr(rec.AltitudeLo),
		Writer:            rec.Writer,
	}
}

func (r *repo) GetSubscription(_ context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	rec, ok := r.state.Subscriptions[id]
	if !ok {
		return nil, nil
	}
	return rec.toModel(), nil
}

func (r *repo) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	if err := validateWriteData(s.Cells, s.StartTime, s.EndTime); err != nil {
		return nil, err
	}
	if _, ok := r.state.Subscriptions[s.ID]; ok {
		return nil, stacktrace.NewError("Subscription with id %s already exists", s.ID)
	}

	now := timestamp.MustFromContext(ctx)

	rec := subRecordFromModel(s, now)
	r.state.Subscriptions[s.ID] = rec
	return rec.toModel(), nil
}

func (r *repo) UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	if err := validateWriteData(s.Cells, s.StartTime, s.EndTime); err != nil {
		return nil, err
	}
	prev, ok := findForWrite(r.state.Subscriptions, s.ID, s.Version)
	if !ok {
		return nil, nil
	}

	now := timestamp.MustFromContext(ctx)

	rec := subRecordFromModel(s, now)
	rec.Owner = prev.Owner // It's not possible to update the owner of a subscription, this ensure it's to changed to a new value.
	r.state.Subscriptions[s.ID] = rec
	return rec.toModel(), nil
}

func (r *repo) DeleteSubscription(_ context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	rec, ok := findForWrite(r.state.Subscriptions, s.ID, s.Version)
	if !ok {
		return nil, nil
	}
	out := rec.toModel()
	delete(r.state.Subscriptions, s.ID)
	return out, nil
}

// liveSubscriptionsInCells yields the non-expired subscriptions touching cells,
// optionally restricted to a single owner.
func (r *repo) liveSubscriptionsInCells(now time.Time, cells s2.CellUnion, owner *dssmodels.Owner) iter.Seq[*subscriptionRecord] {
	return func(yield func(*subscriptionRecord) bool) {

		want := cellSet(cells)
		for _, rec := range r.state.Subscriptions {
			if owner != nil && rec.Owner != *owner {
				continue
			}
			if rec.EndTime == nil || rec.EndTime.Before(now) { // TODO: Don't allow endtime to be null, see #1492
				continue
			}
			if !overlaps(rec.Cells, want) {
				continue
			}
			if !yield(rec) {
				return
			}
		}
	}
}

func (r *repo) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	return r.searchSubscriptions(ctx, cells, nil)
}

func (r *repo) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	return r.searchSubscriptions(ctx, cells, &owner)
}

func (r *repo) searchSubscriptions(ctx context.Context, cells s2.CellUnion, owner *dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "no location provided")
	}

	now := timestamp.MustFromContext(ctx)

	var out []*ridmodels.Subscription
	for rec := range r.liveSubscriptionsInCells(now, cells, owner) {
		out = append(out, rec.toModel())

		if len(out) > dssmodels.MaxResultLimit { // This mimics sqlstore behaviour, but it's not very good.
			break
		}
	}
	return out, nil
}

// UpdateNotificationIdxsInCells increments the notification index for each
// subscription in the given cells.
func (r *repo) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {

	now := timestamp.MustFromContext(ctx)

	var out []*ridmodels.Subscription
	for rec := range r.liveSubscriptionsInCells(now, cells, nil) {
		rec.NotificationIndex++
		out = append(out, rec.toModel())
	}
	return out, nil
}

func (r *repo) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {

	now := timestamp.MustFromContext(ctx)

	want := cellSet(cells)
	counts := make(map[s2.CellID]int, len(cells))
	for rec := range r.liveSubscriptionsInCells(now, cells, &owner) {
		for _, c := range rec.Cells {
			if _, ok := want[c]; ok {
				counts[c]++
			}
		}
	}
	if len(counts) == 0 {
		return 0, nil
	}
	return slices.Max(slices.Collect(maps.Values(counts))), nil
}

func (r *repo) ListExpiredSubscriptions(_ context.Context, writer string, threshold time.Time) ([]*ridmodels.Subscription, error) {
	// TODO: This mimics sqlstore inconsistency of not limiting results there, compared to ISAs. Should it be normalized?
	return listExpired[ridmodels.Subscription](r.state.Subscriptions, writer, threshold, 0), nil
}

func (r *repo) CountSubscriptions(_ context.Context) (int64, error) {
	return int64(len(r.state.Subscriptions)), nil
}
