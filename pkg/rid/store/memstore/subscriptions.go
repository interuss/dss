package memstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
)

func subRecordFromModel(s *ridmodels.Subscription, updatedAt time.Time) *subscriptionRecord {
	return &subscriptionRecord{
		ID:                s.ID,
		URL:               s.URL,
		NotificationIndex: s.NotificationIndex,
		Owner:             s.Owner,
		Cells:             cloneCells(s.Cells),
		StartTime:         cloneTime(s.StartTime),
		EndTime:           cloneTime(s.EndTime),
		AltitudeHi:        cloneFloat32(s.AltitudeHi),
		AltitudeLo:        cloneFloat32(s.AltitudeLo),
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
		Cells:             cloneCells(rec.Cells),
		StartTime:         cloneTime(rec.StartTime),
		EndTime:           cloneTime(rec.EndTime),
		Version:           dssmodels.VersionFromTime(rec.UpdatedAt),
		AltitudeHi:        cloneFloat32(rec.AltitudeHi),
		AltitudeLo:        cloneFloat32(rec.AltitudeLo),
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

func (r *repo) InsertSubscription(_ context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	if err := validateWriteData(s.Cells, s.StartTime, s.EndTime); err != nil {
		return nil, err
	}
	if _, ok := r.state.Subscriptions[s.ID]; ok {
		return nil, stacktrace.NewError("Subscription with id %s already exists", s.ID)
	}
	rec := subRecordFromModel(s, r.clock.Now())
	r.state.Subscriptions[s.ID] = rec
	return rec.toModel(), nil
}

func (r *repo) UpdateSubscription(_ context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	if err := validateWriteData(s.Cells, s.StartTime, s.EndTime); err != nil {
		return nil, err
	}
	prev, ok := r.state.Subscriptions[s.ID]
	if !ok {
		return nil, nil
	}
	if !dssmodels.VersionFromTime(prev.UpdatedAt).Matches(s.Version) {
		return nil, nil
	}
	rec := subRecordFromModel(s, r.clock.Now())
	rec.Owner = prev.Owner
	r.state.Subscriptions[s.ID] = rec
	return rec.toModel(), nil
}

func (r *repo) DeleteSubscription(_ context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	rec, ok := r.state.Subscriptions[s.ID]
	if !ok {
		return nil, nil
	}
	if !dssmodels.VersionFromTime(rec.UpdatedAt).Matches(s.Version) {
		return nil, nil
	}
	out := rec.toModel()
	delete(r.state.Subscriptions, s.ID)
	return out, nil
}

func (r *repo) SearchSubscriptions(_ context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "no location provided")
	}
	now := r.clock.Now()
	want := cellSet(cells)
	var out []*ridmodels.Subscription
	for _, rec := range r.state.Subscriptions {
		if rec.EndTime == nil || rec.EndTime.Before(now) {
			continue
		}
		if !overlaps(rec.Cells, want) {
			continue
		}
		out = append(out, rec.toModel())

		if len(out) > dssmodels.MaxResultLimit { // This miminc sqlstore behaviour, but it's not very good.
			break
		}
	}
	return out, nil
}

func (r *repo) SearchSubscriptionsByOwner(_ context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "no location provided")
	}
	now := r.clock.Now()
	want := cellSet(cells)
	var out []*ridmodels.Subscription
	for _, rec := range r.state.Subscriptions {
		if rec.Owner != owner {
			continue
		}
		if rec.EndTime == nil || rec.EndTime.Before(now) {
			continue
		}
		if !overlaps(rec.Cells, want) {
			continue
		}
		out = append(out, rec.toModel())

		if len(out) > dssmodels.MaxResultLimit { // This miminc sqlstore behaviour, but it's not very good.
			break
		}
	}
	return out, nil
}

// UpdateNotificationIdxsInCells increments the notification index for each
// subscription in the given cells.
func (r *repo) UpdateNotificationIdxsInCells(_ context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	now := r.clock.Now()
	want := cellSet(cells)
	var out []*ridmodels.Subscription
	for _, rec := range r.state.Subscriptions {
		if rec.EndTime == nil || rec.EndTime.Before(now) {
			continue
		}
		if !overlaps(rec.Cells, want) {
			continue
		}
		rec.NotificationIndex++
		out = append(out, rec.toModel())
	}
	return out, nil
}

func (r *repo) MaxSubscriptionCountInCellsByOwner(_ context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	now := r.clock.Now()
	want := cellSet(cells)
	counts := make(map[s2.CellID]int, len(cells))
	for _, rec := range r.state.Subscriptions {
		if rec.Owner != owner {
			continue
		}
		if rec.EndTime == nil || rec.EndTime.Before(now) {
			continue
		}
		for _, c := range rec.Cells {
			if _, ok := want[c]; ok {
				counts[c]++
			}
		}
	}
	best := 0
	for _, n := range counts {
		if n > best {
			best = n
		}
	}
	return best, nil
}

func (r *repo) ListExpiredSubscriptions(_ context.Context, writer string, threshold time.Time) ([]*ridmodels.Subscription, error) {
	var out []*ridmodels.Subscription
	for _, rec := range r.state.Subscriptions {
		// ends_at <= threshold
		if rec.EndTime == nil || rec.EndTime.After(threshold) {
			continue
		}
		if writer == "" {
			if rec.Writer != "" {
				continue
			}
		} else if rec.Writer != writer {
			continue
		}
		out = append(out, rec.toModel())

		// TODO: This miminc sqlstore inconsistency of not limiting results there, comparted to ISAs. Should it be normalized?
	}
	return out, nil
}

func (r *repo) CountSubscriptions(_ context.Context) (int64, error) {
	return int64(len(r.state.Subscriptions)), nil
}
