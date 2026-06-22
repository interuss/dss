package raftstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) SearchConstraints(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Constraint, error) {
	result, err := r.consensus.ProposeValue(ctx, string(searchConstraints), v4d, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose searchConstraints")
	}

	if result == nil {
		return nil, nil
	}

	constraints, ok := result.([]*scdmodels.Constraint)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return constraints, nil
}

func (r *repo) GetConstraint(ctx context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
	result, err := r.consensus.ProposeValue(ctx, string(getConstraint), id, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose getConstraint")
	}

	if result == nil {
		return nil, nil
	}

	constraint, ok := result.(*scdmodels.Constraint)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return constraint, nil
}

func (r *repo) UpsertConstraint(ctx context.Context, constraint *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	result, err := r.consensus.ProposeValue(ctx, string(upsertConstraint), constraint, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose upsertConstraint")
	}

	if result == nil {
		return nil, nil
	}

	upserted, ok := result.(*scdmodels.Constraint)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return upserted, nil
}

func (r *repo) DeleteConstraint(ctx context.Context, id dssmodels.ID) error {
	_, err := r.consensus.ProposeValue(ctx, string(deleteConstraint), id, false)
	return err
}

func (r *repo) CountConstraints(ctx context.Context) (int64, error) {
	result, err := r.consensus.ProposeValue(ctx, string(countConstraints), nil, true)
	if err != nil {
		return 0, stacktrace.Propagate(err, "failed to propose countConstraints")
	}

	if result == nil {
		return 0, nil
	}

	count, ok := result.(int64)
	if !ok {
		return 0, stacktrace.NewError("invalid result type: %T", result)
	}

	return count, nil
}
