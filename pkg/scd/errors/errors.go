package errors

import (
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserrors "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
)

const errMessageMissingOVNs = "Current OVNs not provided for one or more OperationalIntents or Constraints"

var (
	ErrMissingOVNs = stacktrace.NewErrorWithCode(dsserrors.MissingOVNs, errMessageMissingOVNs)
)

// MissingOVNsErrorResponse is Used to return sufficient information for an
// appropriate client error response when a client is missing one or more
// OVNs for relevant OperationalIntents or Constraints.
func MissingOVNsErrorResponse(missingOps []*dssmodels.OperationalIntent, missingConstraints []*dssmodels.Constraint) (*spb.Status, error) {
	detail := &scdpb.AirspaceConflictResponse{
		Message: errMessageMissingOVNs,
	}
	for _, missingOp := range missingOps {
		opRef, err := missingOp.ToProto()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting missing OperationalIntent to proto")
		}
		detail.MissingOperationalIntents = append(detail.MissingOperationalIntents, opRef)
	}
	for _, missingConstraint := range missingConstraints {
		constraintRef, err := missingConstraint.ToProto()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error converting missing Constraint to proto")
		}
		detail.MissingConstraints = append(detail.MissingConstraints, constraintRef)
	}

	p, err := dsserrors.MakeStatusProto(codes.Code(uint16(dsserrors.MissingOVNs)), errMessageMissingOVNs, detail)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error adding AirspaceConflictResponse detail to Status")
	}
	return p, nil
}
