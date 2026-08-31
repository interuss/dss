package actions

import (
	"context"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
)

func init() {
	Registry[restapi.DeleteConstraintReferenceOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.DeleteConstraintReferenceRequest],
		Execute: ExecuteDeleteConstraint,
	}
}

func ExecuteDeleteConstraint(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.DeleteConstraintReferenceRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.DeleteConstraintReferenceOperationID)
	}

	// Retrieve Constraint ID
	id, err := dssmodels.IDFromString(string(req.Entityid))
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format: `%s`", req.Entityid)
	}

	// Make sure deletion request is valid
	old, err := repo.GetConstraint(ctx, id)
	switch {
	case err == pgx.ErrNoRows:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Constraint %s not found", id.String())
	case err != nil:
		return nil, stacktrace.Propagate(err, "Unable to get Constraint from repo")
	case old.Manager != dssmodels.Manager(*req.Auth.ClientID):
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"Constraint owned by %s, but %s attempted to delete", old.Manager, *req.Auth.ClientID)
	case old.OVN != scdmodels.OVN(req.Ovn):
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"Current version is %s but client specified version %s", old.OVN, scdmodels.OVN(req.Ovn))
	}

	// Delete Constraint in repo
	err = repo.DeleteConstraint(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to delete Constraint from repo")
	}

	// Find the Subscriptions interested in Constraints and increment their
	// notification indices.
	subs, err := repo.IncrementNotificationIndicesForConstraints(ctx, &dssmodels.Volume4D{
		StartTime: old.StartTime,
		EndTime:   old.EndTime,
		SpatialVolume: &dssmodels.Volume3D{
			AltitudeHi: old.AltitudeUpper,
			AltitudeLo: old.AltitudeLower,
			Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
				return old.Cells, nil
			}),
		}})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to increment notification indices")
	}

	// Return response to client
	return &restapi.ChangeConstraintReferenceResponse{
		ConstraintReference: *old.ToRest(),
		Subscribers:         makeSubscribersToNotify(subs),
	}, nil
}
