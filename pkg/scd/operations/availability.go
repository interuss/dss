package operations

import (
	"context"

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
	Registry[restapi.GetUssAvailabilityOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:     dssstore.EncodeJSON,
		Decode:     dssstore.DecodeJSON[*restapi.GetUssAvailabilityRequest],
		Execute:    executeGetUssAvailability,
		IsReadOnly: true,
	}
	Registry[restapi.SetUssAvailabilityOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*restapi.SetUssAvailabilityRequest],
		Execute: executeSetUssAvailability,
	}
}

func GetDefaultAvailabilityResponse(id dssmodels.Manager) *restapi.UssAvailabilityStatusResponse {
	return &restapi.UssAvailabilityStatusResponse{
		Status: restapi.UssAvailabilityStatus{
			Availability: restapi.UssAvailabilityState_Unknown,
			Uss:          id.String()},
		Version: "",
	}
}

func executeGetUssAvailability(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.GetUssAvailabilityRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.GetUssAvailabilityOperationID)
	}

	id := dssmodels.ManagerFromString(req.UssId)

	// Get USS availability from Store
	ussa, err := repo.GetUssAvailability(ctx, id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, stacktrace.Propagate(err, "Could not get USS availability from repo")
	}
	if ussa == nil {
		// Return default availability status "Unknown"
		return GetDefaultAvailabilityResponse(id), nil
	}

	return &restapi.UssAvailabilityStatusResponse{
		Status:  *ussa.ToRest(),
		Version: ussa.Version.String(),
	}, nil
}

func executeSetUssAvailability(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	req, ok := request.(*restapi.SetUssAvailabilityRequest)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.SetUssAvailabilityOperationID)
	}

	// Retrieve USS availability status from request params
	availability, err := scdmodels.UssAvailabilityStateFromRest(req.Body.Availability)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid availability state")
	}
	id := dssmodels.ManagerFromString(req.UssId)
	version := scdmodels.OVN(req.Body.OldVersion)
	ussareq := &scdmodels.UssAvailabilityStatus{
		Uss:          id,
		Availability: availability,
	}

	old, err := repo.GetUssAvailability(ctx, id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, stacktrace.Propagate(err, "Could not get USS availability from repo")
	}
	switch {
	case old == nil && !version.Empty():
		// The user wants set a new availability status but it already exists.
		return nil, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "availability for USS %s already exists", id.String())
	case old != nil && old.Version != version:
		// The user wants to update an availability status but the version doesn't match.
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "USS availability version %s is not current", version),
			"Current version is %s but client specified version %s", old.Version, version)
	}

	// Upsert the USS availability
	ussa, err := repo.UpsertUssAvailability(ctx, ussareq)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not upsert USS Availability into repo")
	}
	if ussa == nil {
		return nil, stacktrace.NewError("UpsertUssAvailability returned no USS availability for ID: %s", id)
	}

	return &restapi.UssAvailabilityStatusResponse{
		Status:  *ussa.ToRest(),
		Version: ussa.Version.String(),
	}, nil
}
