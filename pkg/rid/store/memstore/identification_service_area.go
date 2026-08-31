package memstore

import (
	"context"
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

func isaRecordFromModel(isa *ridmodels.IdentificationServiceArea, updatedAt time.Time) *isaRecord {
	return &isaRecord{
		ID:         isa.ID,
		URL:        isa.URL,
		Owner:      isa.Owner,
		Cells:      slices.Clone(isa.Cells),
		StartTime:  utils.ClonePtr(isa.StartTime),
		EndTime:    utils.ClonePtr(isa.EndTime),
		AltitudeHi: utils.ClonePtr(isa.AltitudeHi),
		AltitudeLo: utils.ClonePtr(isa.AltitudeLo),
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
		Cells:      slices.Clone(rec.Cells),
		StartTime:  utils.ClonePtr(rec.StartTime),
		EndTime:    utils.ClonePtr(rec.EndTime),
		Version:    dssmodels.VersionFromTime(rec.UpdatedAt),
		AltitudeHi: utils.ClonePtr(rec.AltitudeHi),
		AltitudeLo: utils.ClonePtr(rec.AltitudeLo),
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

	now := timestamp.MustFromContext(ctx)

	rec := isaRecordFromModel(isa, now)
	r.state.ISAs[isa.ID] = rec
	return rec.toModel(), nil
}

func (r *repo) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	if err := validateWriteData(isa.Cells, isa.StartTime, isa.EndTime); err != nil {
		return nil, err
	}
	prev, ok := findForWrite(r.state.ISAs, isa.ID, isa.Version)
	if !ok {
		return nil, nil
	}

	now := timestamp.MustFromContext(ctx)

	rec := isaRecordFromModel(isa, now)
	rec.Owner = prev.Owner // It's not possible to update the owner of an ISA, this ensure it's to changed to a new value.

	r.state.ISAs[isa.ID] = rec
	return rec.toModel(), nil
}

func (r *repo) DeleteISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	rec, ok := findForWrite(r.state.ISAs, isa.ID, isa.Version)
	if !ok {
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
		if rec.EndTime == nil || rec.EndTime.Before(*earliest) { // TODO: Don't allow endtime to be null, see #1492
			continue
		}
		// COALESCE(starts_at <= latest, true)
		if latest != nil && rec.StartTime != nil && rec.StartTime.After(*latest) { // TODO: Don't allow startup to be null, see #1492
			continue
		}
		if !overlaps(rec.Cells, want) {
			continue
		}
		out = append(out, rec.toModel())

		if len(out) > dssmodels.MaxResultLimit { // This mimics sqlstore behaviour, but it's not very good.
			break
		}
	}
	return out, nil
}

func (r *repo) ListExpiredISAs(_ context.Context, writer string, threshold time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	return listExpired[ridmodels.IdentificationServiceArea](r.state.ISAs, writer, threshold, dssmodels.MaxResultLimit), nil
}

func (r *repo) CountISAs(_ context.Context) (int64, error) {
	return int64(len(r.state.ISAs)), nil
}
