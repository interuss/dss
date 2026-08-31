package raftstore

import (
	"context"
	"encoding/json"
	"strconv"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	raftparams "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
)

type saveOwnMetadataPayload struct {
	Locality       string `json:"locality"`
	PublicEndpoint string `json:"public_endpoint"`
}

func (r *repo) SaveOwnMetadata(ctx context.Context, locality string, publicEndpoint string) error {
	payload := saveOwnMetadataPayload{
		Locality:       locality,
		PublicEndpoint: publicEndpoint,
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal payload")
	}

	_, err = r.consensus.HandleClientRequest(ctx, saveOwnMetadata, buf, false)
	return err
}

func (r *repo) GetDSSMetadata(ctx context.Context) ([]*auxmodels.DSSMetadata, error) {
	result, err := r.consensus.HandleClientRequest(ctx, getDSSMetadata, nil, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to propose %s", getDSSMetadata)
	}

	if res, ok := result.([]*auxmodels.DSSMetadata); ok {
		return res, nil
	}

	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) RecordHeartbeat(ctx context.Context, heartbeat auxmodels.Heartbeat) error {
	buf, err := json.Marshal(heartbeat)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal heartbeat")
	}

	_, err = r.consensus.HandleClientRequest(ctx, recordHeartbeat, buf, false)
	return err
}

func (r *repo) GetDSSAirspaceRepresentationID(_ context.Context) (string, error) {
	connectParameters, err := raftparams.GetConnectParameters("aux")
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to get aux raft parameters")
	}

	return strconv.Itoa(int(connectParameters.ClusterID)), nil
}
