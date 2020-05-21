package errors

import (
	any "github.com/golang/protobuf/ptypes/any"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserrors "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/scd/models"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
	proto "google.golang.org/protobuf/proto"
)

const ErrMessageMissingOVNs = "at least one current OVN not provided"

var (
	errMissingOVNs = status.Error(dsserrors.MissingOVNs, "current OVNS not provided for one or more Operations or Constraints")
)

// Used to return sufficient information for an appropriate client error response when a client is missing one or more
// OVNs for relevant Operations or Constraints.
func MissingOVNsErrorResponse(missingOps []*dssmodels.Operation) (error, bool) {
	response := &scdpb.AirspaceConflictResponse{
		Message: ErrMessageMissingOVNs,
	}
	for _, missingOp := range missingOps {
		opRef, err := missingOp.ToProto()
		if err != nil {
			return err, false
		}
		entityRef := &scdpb.EntityReference{
			OperationReference: opRef,
		}
		response.EntityConflicts = append(response.EntityConflicts, entityRef)
	}

	serialized, err := proto.MarshalOptions{Deterministic: true}.Marshal(response)
	if err != nil {
		return err, false
	}

	p := &spb.Status{
		Code:    int32(dsserrors.MissingOVNs),
		Message: ErrMessageMissingOVNs,
		Details: []*any.Any{
			{
				TypeUrl: "github.com/interuss/dss/" + string(response.ProtoReflect().Descriptor().FullName()),
				Value:   serialized,
			},
		},
	}
	return status.ErrorProto(p), true
}

// A single, consistent error to use internally when the Storage layer detects missing OVNs
func MissingOVNsInternalError() error {
	return errMissingOVNs
}
