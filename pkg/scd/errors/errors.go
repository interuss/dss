package errors

import (
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserrors "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/palantir/stacktrace"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
)

const errMessageMissingOVNs = "At least one current OVN not provided"

var (
	errMissingOVNs = status.Error(dsserrors.MissingOVNs, "Current OVNS not provided for one or more Operations or Constraints")
)

// MissingOVNsErrorResponse is Used to return sufficient information for an
// appropriate client error response when a client is missing one or more
// OVNs for relevant Operations or Constraints.
func MissingOVNsErrorResponse(missingOps []*dssmodels.Operation) (*spb.Status, error) {
	detail := &scdpb.AirspaceConflictResponse{
		Message: errMessageMissingOVNs,
	}
	for _, missingOp := range missingOps {
		opRef, err := missingOp.ToProto()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting missing Operation to proto")
		}
		entityRef := &scdpb.EntityReference{
			OperationReference: opRef,
		}
		detail.EntityConflicts = append(detail.EntityConflicts, entityRef)
	}

	p, err := dsserrors.MakeStatusProto(dsserrors.MissingOVNs, errMessageMissingOVNs, detail)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error adding AirspaceConflictResponse detail to Status")
	}
	return p, nil
}

// MissingOVNsInternalError is a single, consistent error to use internally when
// the Storage layer detects missing OVNs
func MissingOVNsInternalError() error {
	return errMissingOVNs
}
