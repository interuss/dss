package raftstore

import (
	"context"
	"strconv"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/auxpb"
	"github.com/interuss/dss/pkg/pb/raftpb"
	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
)

// proposalPayload wraps an aux payload in the payload of a proposal.
func proposalPayload(payload *auxpb.Payload) *raftpb.Payload {
	return &raftpb.Payload{Payload: &raftpb.Payload_AuxPayload{AuxPayload: payload}}
}

func (r *repo) SaveOwnMetadata(ctx context.Context, locality string, publicEndpoint string) error {
	_, err := r.consensus.HandleWritePayload[any](ctx, proposalPayload(&auxpb.Payload{Payload: &auxpb.Payload_SaveOwnMetadata{
		SaveOwnMetadata: &auxpb.SaveOwnMetadata{Locality: locality, PublicEndpoint: publicEndpoint},
	}}))
	return err
}

func (r *repo) GetDSSMetadata(ctx context.Context) ([]*auxmodels.DSSMetadata, error) {
	return r.consensus.HandleReadPayload[[]*auxmodels.DSSMetadata](ctx, proposalPayload(&auxpb.Payload{Payload: &auxpb.Payload_GetDssMetadata{
		GetDssMetadata: &auxpb.GetDSSMetadata{},
	}}))
}

func (r *repo) RecordHeartbeat(ctx context.Context, heartbeat auxmodels.Heartbeat) error {
	_, err := r.consensus.HandleWritePayload[any](ctx, proposalPayload(&auxpb.Payload{Payload: &auxpb.Payload_RecordHeartbeat{
		RecordHeartbeat: &auxpb.RecordHeartbeat{
			Locality:                    heartbeat.Locality,
			Source:                      heartbeat.Source,
			Timestamp:                   dssmodels.TimeToProto(heartbeat.Timestamp),
			NextHeartbeatExpectedBefore: dssmodels.TimeToProto(heartbeat.NextHeartbeatExpectedBefore),
			Reporter:                    heartbeat.Reporter,
		},
	}}))
	return err
}

func (r *repo) GetDSSAirspaceRepresentationID(_ context.Context) (string, error) {
	connectParameters, err := raftparams.GetConnectParameters("aux")
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to get aux raft parameters")
	}

	return strconv.Itoa(int(connectParameters.ClusterID)), nil
}
