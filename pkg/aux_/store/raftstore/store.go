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

const storeID = "aux_"

const (
	saveOwnMetadata raftstore.RequestType = "saveOwnMetadata"
	getDSSMetadata  raftstore.RequestType = "getDSSMetadata"
	recordHeartbeat raftstore.RequestType = "recordHeartbeat"
)

// repo is a full implementation of aux_.repos.Repository for Raft-based storage.
type repo struct {
	consensus *consensus.Consensus
	memStore  *memstore.Store[repos.Repository]
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

	r := &repo{memStore: memStore}
	store, err := raftstore.Init(ctx, logger, params, r)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize aux raftstore")
	}

	r.consensus = store.Consensus

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) IsReadOnly(requestType raftstore.RequestType) bool {
	return requestType == getDSSMetadata
}

func (r *repo) GetSnapshot() ([]byte, error) {
	return r.memStore.GetSnapshot()
}

func (r *repo) RestoreFromSnapshot(data []byte) error {
	return r.memStore.RestoreFromSnapshot(data)
}

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	mem, err := r.memStore.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to obtain aux memstore repository")
	}

	switch raftstore.RequestType(proposal.RequestType) {
	case saveOwnMetadata:
		var p saveOwnMetadataPayload
		if err := json.Unmarshal(proposal.Value, &p); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", saveOwnMetadata)
		}

		return nil, mem.SaveOwnMetadata(ctx, p.Locality, p.PublicEndpoint)

	case getDSSMetadata:
		return mem.GetDSSMetadata(ctx)

	case recordHeartbeat:
		var hb auxmodels.Heartbeat
		if err := json.Unmarshal(proposal.Value, &hb); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", recordHeartbeat)
		}

		return nil, mem.RecordHeartbeat(ctx, hb)

	default:
		return nil, stacktrace.NewError("unknown request type: %q", proposal.RequestType)
	}
}
