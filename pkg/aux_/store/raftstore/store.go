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
	memStore  *memstore.Store[repos.Repository]
	memRepo   repos.Repository
}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	params, err := auxraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get aux raft parameters")
	}

	memStore, err := auxmemstore.Init(ctx, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize aux memstore")
	}

	r := &repo{memStore: memStore, memRepo: memStore.GetRepo()}
	store, err := raftstore.Init(ctx, logger.With(zap.String("service", "aux_")), params, r, nil)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize aux raftstore")
	}

	r.consensus = store.Consensus

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) GetSnapshot() ([]byte, error) {
	return r.memStore.GetSnapshot()
}

func (r *repo) RestoreFromSnapshot(data []byte) error {
	return r.memStore.RestoreFromSnapshot(data)
}

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	case saveOwnMetadata:
		var payload saveOwnMetadataPayload
		if err := json.Unmarshal(proposal.Value, &payload); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", saveOwnMetadata)
		}

		return nil, r.memRepo.SaveOwnMetadata(ctx, payload.Locality, payload.PublicEndpoint)

	case getDSSMetadata:
		return r.memRepo.GetDSSMetadata(ctx)

	case recordHeartbeat:
		var heartbeat auxmodels.Heartbeat
		if err := json.Unmarshal(proposal.Value, &heartbeat); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", recordHeartbeat)
		}

		return nil, r.memRepo.RecordHeartbeat(ctx, heartbeat)

	default:
		return nil, stacktrace.NewError("unknown request type: %q", proposal.RequestType)
	}
}
