package raftstore

import (
	"context"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) GetOperationalIntent(ctx context.Context, id dssmodels.ID) (*scdmodels.OperationalIntent, error) {
	result, err := r.consensus.ProposeValue(ctx, string(getOperationalIntent), id, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose getOperationalIntent")
	}

	if result == nil {
		return nil, nil
	}

	intent, ok := result.(*scdmodels.OperationalIntent)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return intent, nil
}

func (r *repo) DeleteOperationalIntent(ctx context.Context, id dssmodels.ID) error {
	_, err := r.consensus.ProposeValue(ctx, string(deleteOperationalIntent), id, false)
	return err
}

func (r *repo) UpsertOperationalIntent(ctx context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error) {
	result, err := r.consensus.ProposeValue(ctx, string(upsertOperationalIntent), operation, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose upsertOperationalIntent")
	}

	if result == nil {
		return nil, nil
	}

	intent, ok := result.(*scdmodels.OperationalIntent)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return intent, nil
}

func (r *repo) SearchOperationalIntents(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.OperationalIntent, error) {
	result, err := r.consensus.ProposeValue(ctx, string(searchOperationalIntents), v4d, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose searchOperationalIntents")
	}

	if result == nil {
		return nil, nil
	}

	intents, ok := result.([]*scdmodels.OperationalIntent)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return intents, nil
}

func (r *repo) GetDependentOperationalIntents(ctx context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	result, err := r.consensus.ProposeValue(ctx, string(getDependentOperationalIntents), subscriptionID, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose getDependentOperationalIntents")
	}

	if result == nil {
		return nil, nil
	}

	idList, ok := result.([]dssmodels.ID)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return idList, nil
}

func (r *repo) ListExpiredOperationalIntents(ctx context.Context, threshold time.Time) ([]*scdmodels.OperationalIntent, error) {
	result, err := r.consensus.ProposeValue(ctx, string(listExpiredOperationalIntents), threshold, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose listExpiredOperationalIntents")
	}

	if result == nil {
		return nil, nil
	}

	intents, ok := result.([]*scdmodels.OperationalIntent)
	if !ok {
		return nil, stacktrace.NewError("invalid result type: %T", result)
	}

	return intents, nil
}

func (r *repo) CountOperationalIntents(ctx context.Context) (int64, error) {
	result, err := r.consensus.ProposeValue(ctx, string(countOperationalIntents), nil, true)
	if err != nil {
		return 0, stacktrace.Propagate(err, "failed to propose countOperationalIntents")
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
