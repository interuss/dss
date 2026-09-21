package raftstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang/geo/s2"
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
	listExpiredISAs consensus.RequestType[[]*ridmodels.IdentificationServiceArea] = "listExpiredISAs"
	countISAs       consensus.RequestType[int64]                                  = "countISAs"
)

func (r *repo) GetISA(ctx context.Context, id dssmodels.ID, _ bool) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, getISA, buf, true)
}

func (r *repo) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, deleteISA, buf, false)
}

func (r *repo) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, insertISA, buf, false)
}

func (r *repo) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, updateISA, buf, false)
}

func (r *repo) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(searchISAsPayload{Cells: cells, Earliest: earliest, Latest: latest})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, searchISAs, buf, true)
}

func (r *repo) ListExpiredISAs(ctx context.Context, writer string, threshold time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(expiredPayload{Writer: writer, Threshold: threshold})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, listExpiredISAs, buf, true)
}

func (r *repo) CountISAs(ctx context.Context) (int64, error) {
	return r.consensus.HandleClientRequest(ctx, countISAs, nil, true)
}

func (r *repo) applyISA(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	case string(getISA):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getISA)
		}
		return r.Store.GetRepo().GetISA(ctx, id, false)

	case string(deleteISA):
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteISA)
		}
		return r.Store.GetRepo().DeleteISA(ctx, &isa)

	case string(insertISA):
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", insertISA)
		}
		return r.Store.GetRepo().InsertISA(ctx, &isa)

	case string(updateISA):
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateISA)
		}
		return r.Store.GetRepo().UpdateISA(ctx, &isa)

	case string(searchISAs):
		var payload searchISAsPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchISAs)
		}
		return r.Store.GetRepo().SearchISAs(ctx, payload.Cells, payload.Earliest, payload.Latest)

	case string(listExpiredISAs):
		var payload expiredPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredISAs)
		}
		return r.Store.GetRepo().ListExpiredISAs(ctx, payload.Writer, payload.Threshold)

	case string(countISAs):
		return r.Store.GetRepo().CountISAs(ctx)

	default:
		return nil, stacktrace.NewError("unrecognized ISA request type: %s", proposal.RequestType)
	}
}
