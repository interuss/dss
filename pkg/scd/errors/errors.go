package errors

import (
	proto "github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserrors "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/scd/models"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
)

var (
	errMissingOVNs = status.Error(dsserrors.MissingOVNs, "current OVNS not provided for one or more Operations or Constraints")
)

// Used to return sufficient information for an appropriate client error response when a client is missing one or more
// OVNs for relevant Operations or Constraints.
func MissingOVNsErrorResponse(missingOps []*dssmodels.Operation) (error, error) {
	response := &scdpb.AirspaceConflictResponse{
		Message: "at least one current OVN not provided",
	}
	for _, missingOp := range missingOps {
		opRef, err := missingOp.ToProto()
		if err != nil {
			return nil, err
		}
		entityRef := &scdpb.EntityReference{
			OperationReference: opRef,
		}
		response.EntityConflicts = append(response.EntityConflicts, entityRef)
	}

	serialized, err := proto.Marshal(response)
	if err != nil {
		return nil, err
	}

	p := &spb.Status{
		Code:    int32(dsserrors.MissingOVNs),
		Message: response.Message,
		Details: []*any.Any{
			{
				TypeUrl: "github.com/interuss/dss/" + proto.MessageName(response),
				Value:   serialized,
			},
		},
	}
	return status.ErrorProto(p), nil
}

// A single, consistent error to use internally when the Storage layer detects missing OVNs
func MissingOVNsInternalError() error {
	return errMissingOVNs
}
