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
