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
	getOperationalIntent           consensus.RequestType[*scdmodels.OperationalIntent]   = "getOperationalIntent"
	deleteOperationalIntent        consensus.RequestType[any]                            = "deleteOperationalIntent"
	upsertOperationalIntent        consensus.RequestType[*scdmodels.OperationalIntent]   = "upsertOperationalIntent"
	searchOperationalIntents       consensus.RequestType[[]*scdmodels.OperationalIntent] = "searchOperationalIntents"
	getDependentOperationalIntents consensus.RequestType[[]dssmodels.ID]                 = "getDependentOperationalIntents"
	listExpiredOperationalIntents  consensus.RequestType[[]*scdmodels.OperationalIntent] = "listExpiredOperationalIntents"
	countOperationalIntents        consensus.RequestType[int64]                          = "countOperationalIntents"
)

func (r *repo) GetOperationalIntent(ctx context.Context, id dssmodels.ID) (*scdmodels.OperationalIntent, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, getOperationalIntent, buf, true)
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

	return r.consensus.HandleClientRequest(ctx, upsertOperationalIntent, buf, false)
}

func (r *repo) SearchOperationalIntents(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.OperationalIntent, error) {
	buf, err := json.Marshal(cellsVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, searchOperationalIntents, buf, true)
}

func (r *repo) GetDependentOperationalIntents(ctx context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	buf, err := json.Marshal(subscriptionID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, getDependentOperationalIntents, buf, true)
}

func (r *repo) ListExpiredOperationalIntents(ctx context.Context, threshold time.Time) ([]*scdmodels.OperationalIntent, error) {
	buf, err := json.Marshal(threshold)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, listExpiredOperationalIntents, buf, true)
}

func (r *repo) CountOperationalIntents(ctx context.Context) (int64, error) {
	return r.consensus.HandleClientRequest(ctx, countOperationalIntents, nil, true)
}

func (r *repo) applyOperationalIntent(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	case string(getOperationalIntent):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getOperationalIntent)
		}
		return r.Store.GetRepo().GetOperationalIntent(ctx, id)

	case string(deleteOperationalIntent):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteOperationalIntent)
		}
		return nil, r.Store.GetRepo().DeleteOperationalIntent(ctx, id)

	case string(upsertOperationalIntent):
		var operation scdmodels.OperationalIntent
		if err := json.Unmarshal(proposal.Value, &operation); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", upsertOperationalIntent)
		}
		return r.Store.GetRepo().UpsertOperationalIntent(ctx, &operation)

	case string(searchOperationalIntents):
		var cellsVolume dssmodels.CellsVolume4D
		if err := json.Unmarshal(proposal.Value, &cellsVolume); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchOperationalIntents)
		}
		return r.Store.GetRepo().SearchOperationalIntents(ctx, &cellsVolume)

	case string(getDependentOperationalIntents):
		var subscriptionID dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &subscriptionID); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getDependentOperationalIntents)
		}
		return r.Store.GetRepo().GetDependentOperationalIntents(ctx, subscriptionID)

	case string(listExpiredOperationalIntents):
		var threshold time.Time
		if err := json.Unmarshal(proposal.Value, &threshold); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredOperationalIntents)
		}
		return r.Store.GetRepo().ListExpiredOperationalIntents(ctx, threshold)

	case string(countOperationalIntents):
		return r.Store.GetRepo().CountOperationalIntents(ctx)

	default:
		return nil, stacktrace.NewError("unrecognized operational intent request type: %s", proposal.RequestType)
	}
}
