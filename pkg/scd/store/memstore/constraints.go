package memstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	dsssql "github.com/interuss/dss/pkg/sql"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
)

func (rec *constraintRecord) toModel() *scdmodels.Constraint {
	return &scdmodels.Constraint{
		ID:            rec.ID,
		Manager:       rec.Manager,
		Version:       rec.Version,
		OVN:           scdmodels.NewOVNFromTime(rec.UpdatedAt, rec.ID.String()),
		StartTime:     cloneTime(rec.StartTime),
		EndTime:       cloneTime(rec.EndTime),
		USSBaseURL:    rec.USSBaseURL,
		AltitudeLower: cloneFloat32(rec.AltitudeLower),
		AltitudeUpper: cloneFloat32(rec.AltitudeUpper),
		Cells:         cloneCells(rec.Cells),
	}
}

// SearchConstraints implements scd.repos.Constraint.SearchConstraints.
func (r *repo) SearchConstraints(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Constraint, error) {
	cells, err := v4d.CalculateSpatialCovering()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not calculate spatial covering")
	}
	if len(cells) == 0 {
		return []*scdmodels.Constraint{}, nil
	}

	want := cellSet(cells)
	var out []*scdmodels.Constraint
	for _, rec := range r.state.Constraints {
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

// GetConstraint implements scd.repos.Constraint.GetConstraint.
func (r *repo) GetConstraint(_ context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
	rec, ok := r.state.Constraints[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return rec.toModel(), nil
}

// UpsertConstraint implements scd.repos.Constraint.UpsertConstraint.
func (r *repo) UpsertConstraint(ctx context.Context, s *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	if _, err := dsssql.CellUnionToCellIdsWithValidation(s.Cells); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}

	rec := &constraintRecord{
		ID:            s.ID,
		Manager:       s.Manager,
		Version:       s.Version,
		StartTime:     cloneTime(s.StartTime),
		EndTime:       cloneTime(s.EndTime),
		USSBaseURL:    s.USSBaseURL,
		AltitudeLower: cloneFloat32(s.AltitudeLower),
		AltitudeUpper: cloneFloat32(s.AltitudeUpper),
		Cells:         cloneCells(s.Cells),
		UpdatedAt:     timestamp.NowFromContext(ctx),
	}
	r.state.Constraints[s.ID] = rec
	return rec.toModel(), nil
}

// DeleteConstraint implements scd.repos.Constraint.DeleteConstraint.
func (r *repo) DeleteConstraint(_ context.Context, id dssmodels.ID) error {
	if _, ok := r.state.Constraints[id]; !ok {
		return pgx.ErrNoRows
	}
	delete(r.state.Constraints, id)
	return nil
}

func (r *repo) CountConstraints(_ context.Context) (int64, error) {
	return int64(len(r.state.Constraints)), nil
}
