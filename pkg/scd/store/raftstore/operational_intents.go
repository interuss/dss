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

	return r.consensus.HandleReadRequest(ctx, getOperationalIntent, buf)
}

func (r *repo) DeleteOperationalIntent(ctx context.Context, id dssmodels.ID) error {
	buf, err := json.Marshal(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal payload")
	}

	_, err = r.consensus.HandleWriteRequest(ctx, deleteOperationalIntent, buf)
	return err
}

func (r *repo) UpsertOperationalIntent(ctx context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error) {
	buf, err := json.Marshal(operation)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleWriteRequest(ctx, upsertOperationalIntent, buf)
}

func (r *repo) SearchOperationalIntents(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.OperationalIntent, error) {
	buf, err := json.Marshal(cellsVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, searchOperationalIntents, buf)
}

func (r *repo) GetDependentOperationalIntents(ctx context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	buf, err := json.Marshal(subscriptionID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, getDependentOperationalIntents, buf)
}

func (r *repo) ListExpiredOperationalIntents(ctx context.Context, threshold time.Time) ([]*scdmodels.OperationalIntent, error) {
	buf, err := json.Marshal(threshold)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, listExpiredOperationalIntents, buf)
}

func (r *repo) CountOperationalIntents(ctx context.Context) (int64, error) {
	return r.consensus.HandleReadRequest(ctx, countOperationalIntents, nil)
}
