package raftstore

import (
	"context"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	"github.com/interuss/dss/pkg/aux_/repos"
	auxmemstore "github.com/interuss/dss/pkg/aux_/store/memstore"
	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/auxpb"
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of aux_.repos.Repository for Raft-based storage.
type repo struct {
	consensus *consensus.Consensus
	*memstore.Store[repos.Repository]
}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	params, err := raftparams.GetConnectParameters("aux")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get aux raft parameters")
	}

	memStore, err := auxmemstore.Init(ctx, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize aux memstore")
	}

	r := &repo{Store: memStore}
	store, err := raftstore.Init(ctx, logger.With(zap.String("service", "aux_")), params, r, nil)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize aux raftstore")
	}

	r.consensus = store.Consensus

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch req := proposal.Payload.GetAuxPayload().GetPayload().(type) {
	case *auxpb.Payload_SaveOwnMetadata:
		metadata := req.SaveOwnMetadata
		return nil, r.Store.GetRepo().SaveOwnMetadata(ctx, metadata.GetLocality(), metadata.GetPublicEndpoint())

	case *auxpb.Payload_GetDssMetadata:
		return r.Store.GetRepo().GetDSSMetadata(ctx)

	case *auxpb.Payload_RecordHeartbeat:
		heartbeat := req.RecordHeartbeat
		return nil, r.Store.GetRepo().RecordHeartbeat(ctx, auxmodels.Heartbeat{
			Locality:                    heartbeat.GetLocality(),
			Source:                      heartbeat.GetSource(),
			Timestamp:                   dssmodels.TimeFromProto(heartbeat.GetTimestamp()),
			NextHeartbeatExpectedBefore: dssmodels.TimeFromProto(heartbeat.GetNextHeartbeatExpectedBefore()),
			Reporter:                    heartbeat.GetReporter(),
		})

	default:
		return nil, stacktrace.NewError("unknown request: %q", proposal.Name())
	}
}
