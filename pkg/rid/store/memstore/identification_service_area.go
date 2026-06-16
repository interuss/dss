package memstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

func isaRecordFromModel(isa *ridmodels.IdentificationServiceArea, updatedAt time.Time) *isaRecord {
	return &isaRecord{
		ID:         isa.ID,
		URL:        isa.URL,
		Owner:      isa.Owner,
		Cells:      cloneCells(isa.Cells),
		StartTime:  cloneTime(isa.StartTime),
		EndTime:    cloneTime(isa.EndTime),
		AltitudeHi: cloneFloat32(isa.AltitudeHi),
		AltitudeLo: cloneFloat32(isa.AltitudeLo),
		Writer:     isa.Writer,
		UpdatedAt:  updatedAt,
	}
}

// toModel rebuilds the ISA model
func (rec *isaRecord) toModel() *ridmodels.IdentificationServiceArea {
	return &ridmodels.IdentificationServiceArea{
		ID:         rec.ID,
		URL:        rec.URL,
		Owner:      rec.Owner,
		Cells:      cloneCells(rec.Cells),
		StartTime:  cloneTime(rec.StartTime),
		EndTime:    cloneTime(rec.EndTime),
		Version:    dssmodels.VersionFromTime(rec.UpdatedAt),
		AltitudeHi: cloneFloat32(rec.AltitudeHi),
		AltitudeLo: cloneFloat32(rec.AltitudeLo),
		Writer:     rec.Writer,
	}
}

func (r *repo) GetISA(_ context.Context, id dssmodels.ID, _ bool) (*ridmodels.IdentificationServiceArea, error) {
	rec, ok := r.state.ISAs[id]
	if !ok {
		return nil, nil
	}
	return rec.toModel(), nil
}

func (r *repo) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	if err := validateWriteData(isa.Cells, isa.StartTime, isa.EndTime); err != nil {
		return nil, err
	}
	if _, ok := r.state.ISAs[isa.ID]; ok {
		return nil, stacktrace.NewError("ISA with id %s already exists", isa.ID)
	}
	rec := isaRecordFromModel(isa, timestamp.NowFromContext(ctx))
	r.state.ISAs[isa.ID] = rec
	return rec.toModel(), nil
}

func (r *repo) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	if err := validateWriteData(isa.Cells, isa.StartTime, isa.EndTime); err != nil {
		return nil, err
	}
	prev, ok := r.state.ISAs[isa.ID]
	if !ok {
		return nil, nil
	}
	if !dssmodels.VersionFromTime(prev.UpdatedAt).Matches(isa.Version) {
		return nil, nil
	}
	rec := isaRecordFromModel(isa, timestamp.NowFromContext(ctx))
	rec.Owner = prev.Owner
	r.state.ISAs[isa.ID] = rec
	return rec.toModel(), nil
}

func (r *repo) DeleteISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	rec, ok := r.state.ISAs[isa.ID]
	if !ok {
		return nil, nil
	}
	if !dssmodels.VersionFromTime(rec.UpdatedAt).Matches(isa.Version) {
		return nil, nil
	}
	out := rec.toModel()
	delete(r.state.ISAs, isa.ID)
	return out, nil
}

func (r *repo) SearchISAs(_ context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	if len(cells) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing cell IDs for query")
	}
	if earliest == nil {
		return nil, stacktrace.NewError("Earliest start time is missing")
	}

	want := cellSet(cells)
	var out []*ridmodels.IdentificationServiceArea
	for _, rec := range r.state.ISAs {
		// ends_at >= earliest
		if rec.EndTime == nil || rec.EndTime.Before(*earliest) {
			continue
		}
		// COALESCE(starts_at <= latest, true)
		if latest != nil && rec.StartTime != nil && rec.StartTime.After(*latest) {
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

func (r *repo) ListExpiredISAs(_ context.Context, writer string, threshold time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	var out []*ridmodels.IdentificationServiceArea
	for _, rec := range r.state.ISAs {
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

		if len(out) > dssmodels.MaxResultLimit { // This miminc sqlstore behaviour, but it's not very good.
			break
		}
	}
	return out, nil
}

func (r *repo) CountISAs(_ context.Context) (int64, error) {
	return int64(len(r.state.ISAs)), nil
}
