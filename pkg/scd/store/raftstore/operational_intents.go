package raftstore

import (
	"context"
	"encoding/json"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

const (
	getOperationalIntent           consensus.RequestType = "getOperationalIntent"
	deleteOperationalIntent        consensus.RequestType = "deleteOperationalIntent"
	upsertOperationalIntent        consensus.RequestType = "upsertOperationalIntent"
	searchOperationalIntents       consensus.RequestType = "searchOperationalIntents"
	getDependentOperationalIntents consensus.RequestType = "getDependentOperationalIntents"
	listExpiredOperationalIntents  consensus.RequestType = "listExpiredOperationalIntents"
	countOperationalIntents        consensus.RequestType = "countOperationalIntents"
)

func (r *repo) GetOperationalIntent(ctx context.Context, id dssmodels.ID) (*scdmodels.OperationalIntent, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, getOperationalIntent, buf, true)
	if err != nil {
		return nil, err
	}
	if operation, ok := result.(*scdmodels.OperationalIntent); ok {
		return operation, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) DeleteOperationalIntent(ctx context.Context, id dssmodels.ID) error {
	buf, err := json.Marshal(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal payload")
	}

	_, err = r.consensus.HandleClientRequest(ctx, deleteOperationalIntent, buf, false)
	return err
}

func (r *repo) UpsertOperationalIntent(ctx context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error) {
	buf, err := json.Marshal(operation)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, upsertOperationalIntent, buf, false)
	if err != nil {
		return nil, err
	}
	if upserted, ok := result.(*scdmodels.OperationalIntent); ok {
		return upserted, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) SearchOperationalIntents(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.OperationalIntent, error) {
	buf, err := json.Marshal(v4d)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, searchOperationalIntents, buf, true)
	if err != nil {
		return nil, err
	}
	if operations, ok := result.([]*scdmodels.OperationalIntent); ok {
		return operations, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) GetDependentOperationalIntents(ctx context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	buf, err := json.Marshal(subscriptionID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, getDependentOperationalIntents, buf, true)
	if err != nil {
		return nil, err
	}
	if ids, ok := result.([]dssmodels.ID); ok {
		return ids, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) ListExpiredOperationalIntents(ctx context.Context, threshold time.Time) ([]*scdmodels.OperationalIntent, error) {
	buf, err := json.Marshal(threshold)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, listExpiredOperationalIntents, buf, true)
	if err != nil {
		return nil, err
	}
	if operations, ok := result.([]*scdmodels.OperationalIntent); ok {
		return operations, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) CountOperationalIntents(ctx context.Context) (int64, error) {
	result, err := r.consensus.HandleClientRequest(ctx, countOperationalIntents, nil, true)
	if err != nil {
		return 0, err
	}
	if count, ok := result.(int64); ok {
		return count, nil
	}
	return 0, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) applyOperationalIntent(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	case getOperationalIntent:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getOperationalIntent)
		}
		return r.Store.GetRepo().GetOperationalIntent(ctx, id)

	case deleteOperationalIntent:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteOperationalIntent)
		}
		return nil, r.Store.GetRepo().DeleteOperationalIntent(ctx, id)

	case upsertOperationalIntent:
		var operation scdmodels.OperationalIntent
		if err := json.Unmarshal(proposal.Value, &operation); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", upsertOperationalIntent)
		}
		return r.Store.GetRepo().UpsertOperationalIntent(ctx, &operation)

	case searchOperationalIntents:
		var v4d dssmodels.Volume4D
		if err := json.Unmarshal(proposal.Value, &v4d); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchOperationalIntents)
		}
		return r.Store.GetRepo().SearchOperationalIntents(ctx, &v4d)

	case getDependentOperationalIntents:
		var subscriptionID dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &subscriptionID); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getDependentOperationalIntents)
		}
		return r.Store.GetRepo().GetDependentOperationalIntents(ctx, subscriptionID)

	case listExpiredOperationalIntents:
		var threshold time.Time
		if err := json.Unmarshal(proposal.Value, &threshold); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredOperationalIntents)
		}
		return r.Store.GetRepo().ListExpiredOperationalIntents(ctx, threshold)

	case countOperationalIntents:
		return r.Store.GetRepo().CountOperationalIntents(ctx)

	default:
		return nil, stacktrace.NewError("unrecognized operational intent request type: %s", proposal.RequestType)
	}
}
