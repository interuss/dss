package memstore

import (
	"context"
	"slices"

	"github.com/interuss/dss/pkg/memstore/utils"
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
		StartTime:     utils.ClonePtr(rec.StartTime),
		EndTime:       utils.ClonePtr(rec.EndTime),
		USSBaseURL:    rec.USSBaseURL,
		AltitudeLower: utils.ClonePtr(rec.AltitudeLower),
		AltitudeUpper: utils.ClonePtr(rec.AltitudeUpper),
		Cells:         slices.Clone(rec.Cells),
	}
}

func (r *repo) SearchConstraints(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Constraint, error) {
	want, err := coveringSet(v4d)
	if err != nil {
		return nil, err
	}
	if len(want) == 0 {
		return []*scdmodels.Constraint{}, nil
	}

	var out []*scdmodels.Constraint
	for _, rec := range r.state.Constraints {
		if !overlaps(rec.Cells, want) {
			continue
		}
		if !overlapsTime(rec.StartTime, rec.EndTime, v4d) {
			continue
		}
		out = append(out, rec.toModel())

		if len(out) >= dssmodels.MaxResultLimit { // mirror SQL "LIMIT MaxResultLimit"
			break
		}
	}
	return out, nil
}

func (r *repo) GetConstraint(_ context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
	rec, ok := r.state.Constraints[id]
	if !ok {
		return nil, pgx.ErrNoRows // TODO: #1608
	}
	return rec.toModel(), nil
}

func (r *repo) UpsertConstraint(ctx context.Context, s *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	if _, err := dsssql.CellUnionToCellIdsWithValidation(s.Cells); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert array to jackc/pgtype")
	}

	now := timestamp.MustFromContext(ctx)

	rec := &constraintRecord{
		ID:            s.ID,
		Manager:       s.Manager,
		Version:       s.Version,
		StartTime:     utils.ClonePtr(s.StartTime),
		EndTime:       utils.ClonePtr(s.EndTime),
		USSBaseURL:    s.USSBaseURL,
		AltitudeLower: utils.ClonePtr(s.AltitudeLower),
		AltitudeUpper: utils.ClonePtr(s.AltitudeUpper),
		Cells:         slices.Clone(s.Cells),
		UpdatedAt:     now,
	}
	r.state.Constraints[s.ID] = rec
	return rec.toModel(), nil
}

func (r *repo) DeleteConstraint(_ context.Context, id dssmodels.ID) error {
	if _, ok := r.state.Constraints[id]; !ok {
		return pgx.ErrNoRows // TODO: #1608
	}
	delete(r.state.Constraints, id)
	return nil
}

func (r *repo) CountConstraints(_ context.Context) (int64, error) {
	return int64(len(r.state.Constraints)), nil
}
