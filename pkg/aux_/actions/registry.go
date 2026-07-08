package actions

import (
	"context"
	"encoding/json"

	restapi "github.com/interuss/dss/pkg/api/auxv1"
	"github.com/interuss/dss/pkg/aux_/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
)

// Registry maps operation IDs to their handlers.
var Registry = map[string]dssstore.OperationHandler[repos.Repository]{
	restapi.GetDSSInstancesOperationID: {
		Encode: encodeRequest,
		Decode: decodeGetDSSInstances,
		Execute: func(ctx context.Context, r repos.Repository, req dssstore.OperationRequest) (any, error) {
			return GetDSSInstances(ctx, r, req.(*restapi.GetDSSInstancesRequest))
		},
		IsReadOnly: true,
	},
	restapi.PutDSSInstancesHeartbeatOperationID: {
		Encode: encodeRequest,
		Decode: decodePutDSSInstancesHeartbeat,
		Execute: func(ctx context.Context, r repos.Repository, req dssstore.OperationRequest) (any, error) {
			return PutDSSInstancesHeartbeat(ctx, r, req.(*restapi.PutDSSInstancesHeartbeatRequest))
		},
		IsReadOnly: false,
	},
}

func encodeRequest(req dssstore.OperationRequest) ([]byte, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to encode op %q", req.OperationID())
	}
	return data, nil
}

func decodeGetDSSInstances(buf []byte) (dssstore.OperationRequest, error) {
	var req restapi.GetDSSInstancesRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode %s", restapi.GetDSSInstancesOperationID)
	}
	return &req, nil
}

func decodePutDSSInstancesHeartbeat(buf []byte) (dssstore.OperationRequest, error) {
	var req restapi.PutDSSInstancesHeartbeatRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode %s", restapi.PutDSSInstancesHeartbeatOperationID)
	}
	return &req, nil
}
