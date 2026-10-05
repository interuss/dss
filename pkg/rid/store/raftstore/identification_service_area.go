package raftstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
)

const (
	getISA          consensus.RequestType[*ridmodels.IdentificationServiceArea]   = "getISA"
	deleteISA       consensus.RequestType[*ridmodels.IdentificationServiceArea]   = "deleteISA"
	insertISA       consensus.RequestType[*ridmodels.IdentificationServiceArea]   = "insertISA"
	updateISA       consensus.RequestType[*ridmodels.IdentificationServiceArea]   = "updateISA"
	searchISAs      consensus.RequestType[[]*ridmodels.IdentificationServiceArea] = "searchISAs"
	listExpiredISAs consensus.RequestType[[]dssmodels.ID]                         = "listExpiredISAs"
	countISAs       consensus.RequestType[int64]                                  = "countISAs"
)

func (r *repo) GetISA(ctx context.Context, id dssmodels.ID, _ bool) (*ridmodels.IdentificationServiceArea, error) {
	// We ignore the forUpdate parameter because raft requests are executed sequentially, so there is no need to lock the ISA for update.

	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, getISA, buf)
}

func (r *repo) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleWriteRequest(ctx, deleteISA, buf)
}

func (r *repo) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleWriteRequest(ctx, insertISA, buf)
}

func (r *repo) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleWriteRequest(ctx, updateISA, buf)
}

func (r *repo) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(searchISAsPayload{Cells: cells, Earliest: earliest, Latest: latest})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, searchISAs, buf)
}

func (r *repo) ListExpiredISAs(ctx context.Context, writer string, threshold time.Time) ([]dssmodels.ID, error) {
	buf, err := json.Marshal(expiredPayload{Writer: writer, Threshold: threshold})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleReadRequest(ctx, listExpiredISAs, buf)
}

func (r *repo) DeleteExpiredISAs(_ context.Context, writer string, threshold time.Time) ([]dssmodels.ID, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteExpiredISAs not implemented for raftstore")
}

func (r *repo) CountISAs(ctx context.Context) (int64, error) {
	return r.consensus.HandleReadRequest(ctx, countISAs, nil)
}
