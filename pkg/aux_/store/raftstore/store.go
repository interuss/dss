package raftstore

import (
	"context"
	"encoding/json"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	"github.com/interuss/dss/pkg/aux_/repos"
	auxmemstore "github.com/interuss/dss/pkg/aux_/store/memstore"
	auxraftparams "github.com/interuss/dss/pkg/aux_/store/raftstore/params"
	"github.com/interuss/dss/pkg/memstore"
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

const (
	saveOwnMetadata consensus.RequestType = "saveOwnMetadata"
	getDSSMetadata  consensus.RequestType = "getDSSMetadata"
	recordHeartbeat consensus.RequestType = "recordHeartbeat"
)

// repo is a full implementation of aux_.repos.Repository for Raft-based storage.
type repo struct {
	consensus *consensus.Consensus
	*memstore.Store[repos.Repository]
}

func Init(ctx context.Context, logger *zap.Logger, locality string) (*raftstore.Store[repos.Repository], error) {
	params, err := auxraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get aux raft parameters")
	}

	memStore, err := auxmemstore.Init(ctx, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize aux memstore")
	}

	r := &repo{Store: memStore}
	store, err := raftstore.Init(ctx, logger.With(zap.String("service", "aux_")), locality, params, r, nil)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize aux raftstore")
	}

	r.consensus = store.Consensus

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	case saveOwnMetadata:
		var payload saveOwnMetadataPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", saveOwnMetadata)
		}

		return nil, r.Store.GetRepo().SaveOwnMetadata(ctx, payload.Locality, payload.PublicEndpoint)

	case getDSSMetadata:
		return r.Store.GetRepo().GetDSSMetadata(ctx)

	case recordHeartbeat:
		var heartbeat auxmodels.Heartbeat
		if err := json.Unmarshal(proposal.Value, &heartbeat); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", recordHeartbeat)
		}

		return nil, r.Store.GetRepo().RecordHeartbeat(ctx, heartbeat)

	default:
		return nil, stacktrace.NewError("unknown request type: %q", proposal.RequestType)
	}
}
