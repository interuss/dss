package raftstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) GetUssAvailability(ctx context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	result, err := r.consensus.ProposeValue(ctx, string(getUSSAvailability), id, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose getUSSAvailability")
	}

	if result == nil {
		return nil, nil
	}

	status, ok := result.(*scdmodels.UssAvailabilityStatus)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return status, nil
}

func (r *repo) UpsertUssAvailability(ctx context.Context, ussa *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	result, err := r.consensus.ProposeValue(ctx, string(upsertUSSAvailability), ussa, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose upsertUSSAvailability")
	}

	if result == nil {
		return nil, nil
	}

	status, ok := result.(*scdmodels.UssAvailabilityStatus)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return status, nil
}
