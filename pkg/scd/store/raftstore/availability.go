package raftstore

import (
	"context"
	"encoding/json"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

const (
	getUssAvailability    consensus.RequestType = "getUssAvailability"
	upsertUssAvailability consensus.RequestType = "upsertUssAvailability"
)

func (r *repo) GetUssAvailability(ctx context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, getUssAvailability, buf, true)
	if err != nil {
		return nil, err
	}
	if ussa, ok := result.(*scdmodels.UssAvailabilityStatus); ok {
		return ussa, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) UpsertUssAvailability(ctx context.Context, ussa *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	buf, err := json.Marshal(ussa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, upsertUssAvailability, buf, false)
	if err != nil {
		return nil, err
	}
	if upserted, ok := result.(*scdmodels.UssAvailabilityStatus); ok {
		return upserted, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) applyAvailability(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	case getUssAvailability:
		var manager dssmodels.Manager
		if err := json.Unmarshal(proposal.Value, &manager); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getUssAvailability)
		}
		return r.Store.GetRepo().GetUssAvailability(ctx, manager)

	case upsertUssAvailability:
		var ussa scdmodels.UssAvailabilityStatus
		if err := json.Unmarshal(proposal.Value, &ussa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", upsertUssAvailability)
		}
		return r.Store.GetRepo().UpsertUssAvailability(ctx, &ussa)

	default:
		return nil, stacktrace.NewError("unrecognized availability request type: %s", proposal.RequestType)
	}
}
