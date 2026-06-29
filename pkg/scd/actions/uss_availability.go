package actions

import (
	"context"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5"
)

type GetUssAvailabilityAction struct {
	ID dssmodels.Manager
}

func (a *GetUssAvailabilityAction) RequestType() string { return "getUssAvailability" }

func (a *GetUssAvailabilityAction) IsReadOnly() bool { return true }

func (a *GetUssAvailabilityAction) Execute(ctx context.Context, r repos.Repository) (any, error) {
	// Get USS availability from Store
	ussa, err := r.GetUssAvailability(ctx, a.ID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, stacktrace.Propagate(err, "Could not get USS availability from repo")
	}
	if ussa == nil {
		// Return default availability status "Unknown"
		return GetDefaultAvailabilityResponse(a.ID), nil
	}
	return &restapi.UssAvailabilityStatusResponse{
		Status:  *ussa.ToRest(),
		Version: ussa.Version.String(),
	}, nil
}

type SetUssAvailabilityAction struct {
	ID           dssmodels.Manager
	Version      scdmodels.OVN
	Availability scdmodels.UssAvailabilityState
}

func (a *SetUssAvailabilityAction) RequestType() string { return "setUssAvailability" }

func (a *SetUssAvailabilityAction) IsReadOnly() bool { return false }

func (a *SetUssAvailabilityAction) Execute(ctx context.Context, r repos.Repository) (any, error) {
	old, err := r.GetUssAvailability(ctx, a.ID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, stacktrace.Propagate(err, "Could not get USS availability from repo")
	}
	switch {
	case old == nil && !a.Version.Empty():
		// The user wants set a new availability status but it already exists.
		return nil, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "availability for USS %s already exists", a.ID.String())
	case old != nil && old.Version != a.Version:
		// The user wants to update an availability status but the version doesn't match.
		return nil, stacktrace.Propagate(
			stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "USS availability version %s is not current", a.Version),
			"Current version is %s but client specified version %s", old.Version, a.Version)
	}

	// Upsert the USS availability
	ussa, err := r.UpsertUssAvailability(ctx, &scdmodels.UssAvailabilityStatus{
		Uss:          a.ID,
		Availability: a.Availability,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not upsert USS Availability into repo")
	}
	if ussa == nil {
		return nil, stacktrace.NewError("UpsertUssAvailability returned no USS availability for ID: %s", a.ID)
	}

	return &restapi.UssAvailabilityStatusResponse{
		Status:  *ussa.ToRest(),
		Version: ussa.Version.String(),
	}, nil
}

func GetDefaultAvailabilityResponse(id dssmodels.Manager) *restapi.UssAvailabilityStatusResponse {
	return &restapi.UssAvailabilityStatusResponse{
		Status: restapi.UssAvailabilityStatus{
			Availability: restapi.UssAvailabilityState_Unknown,
			Uss:          id.String()},
		Version: "",
	}
}
