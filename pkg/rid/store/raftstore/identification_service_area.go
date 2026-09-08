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
	getISA          consensus.RequestType = "getISA"
	deleteISA       consensus.RequestType = "deleteISA"
	insertISA       consensus.RequestType = "insertISA"
	updateISA       consensus.RequestType = "updateISA"
	searchISAs      consensus.RequestType = "searchISAs"
	listExpiredISAs consensus.RequestType = "listExpiredISAs"
	countISAs       consensus.RequestType = "countISAs"
)

func (r *repo) GetISA(ctx context.Context, id dssmodels.ID, _ bool) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, getISA, buf, true)
	if err != nil {
		return nil, err
	}
	if isa, ok := result.(*ridmodels.IdentificationServiceArea); ok {
		return isa, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, deleteISA, buf, false)
	if err != nil {
		return nil, err
	}
	if deleted, ok := result.(*ridmodels.IdentificationServiceArea); ok {
		return deleted, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, insertISA, buf, false)
	if err != nil {
		return nil, err
	}
	if inserted, ok := result.(*ridmodels.IdentificationServiceArea); ok {
		return inserted, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, updateISA, buf, false)
	if err != nil {
		return nil, err
	}
	if updated, ok := result.(*ridmodels.IdentificationServiceArea); ok {
		return updated, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(searchISAsPayload{Cells: cells, Earliest: earliest, Latest: latest})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, searchISAs, buf, true)
	if err != nil {
		return nil, err
	}
	if isas, ok := result.([]*ridmodels.IdentificationServiceArea); ok {
		return isas, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) ListExpiredISAs(ctx context.Context, writer string, threshold time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	buf, err := json.Marshal(expiredPayload{Writer: writer, Threshold: threshold})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, listExpiredISAs, buf, true)
	if err != nil {
		return nil, err
	}
	if isas, ok := result.([]*ridmodels.IdentificationServiceArea); ok {
		return isas, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) CountISAs(ctx context.Context) (int64, error) {
	result, err := r.consensus.HandleClientRequest(ctx, countISAs, nil, true)
	if err != nil {
		return 0, err
	}
	if count, ok := result.(int64); ok {
		return count, nil
	}
	return 0, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) applyISA(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	case getISA:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getISA)
		}
		return r.Store.GetRepo().GetISA(ctx, id, false)

	case deleteISA:
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteISA)
		}
		return r.Store.GetRepo().DeleteISA(ctx, &isa)

	case insertISA:
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", insertISA)
		}
		return r.Store.GetRepo().InsertISA(ctx, &isa)

	case updateISA:
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateISA)
		}
		return r.Store.GetRepo().UpdateISA(ctx, &isa)

	case searchISAs:
		var payload searchISAsPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchISAs)
		}
		return r.Store.GetRepo().SearchISAs(ctx, payload.Cells, payload.Earliest, payload.Latest)

	case listExpiredISAs:
		var payload expiredPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredISAs)
		}
		return r.Store.GetRepo().ListExpiredISAs(ctx, payload.Writer, payload.Threshold)

	case countISAs:
		return r.Store.GetRepo().CountISAs(ctx)

	default:
		return nil, stacktrace.NewError("unrecognized ISA request type: %s", proposal.RequestType)
	}
}
