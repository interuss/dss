package operations

import (
	"context"
	"errors"

	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/dss/pkg/pb/scdpb"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/stacktrace"
)

func init() {
	Registry[restapi.GetUssAvailabilityOperationID] = dssstore.OperationHandler[repos.Repository]{
		EncodePayload: encodeGetUssAvailability,
		DecodePayload: decodeGetUssAvailability,
		Execute:       executeGetUssAvailability,
		IsReadOnly:    true,
	}
	Registry[restapi.SetUssAvailabilityOperationID] = dssstore.OperationHandler[repos.Repository]{
		EncodePayload: encodeSetUssAvailability,
		DecodePayload: decodeSetUssAvailability,
		Execute:       executeSetUssAvailability,
	}
}

type getUssAvailabilityPayload struct {
	Uss dssmodels.Manager
}

func (p *getUssAvailabilityPayload) OperationID() string {
	return restapi.GetUssAvailabilityOperationID
}

// NewGetUssAvailabilityPayload builds the request for a USS availability retrieval.
func NewGetUssAvailabilityPayload(uss dssmodels.Manager) dssstore.OperationRequest {
	return &getUssAvailabilityPayload{Uss: uss}
}

func encodeGetUssAvailability(request dssstore.OperationRequest) (*raftpb.Payload, error) {
	payload, ok := request.(*getUssAvailabilityPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.GetUssAvailabilityOperationID)
	}
	return ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_GetUssAvailabilityOperation{
		GetUssAvailabilityOperation: &scdpb.GetUssAvailabilityOperation{Uss: payload.Uss.String()},
	}}), nil
}

func decodeGetUssAvailability(payload *raftpb.Payload) (dssstore.OperationRequest, error) {
	req := payload.GetScdPayload().GetGetUssAvailabilityOperation()
	if req == nil {
		return nil, stacktrace.NewError("unexpected payload %T for operation %q", payload.GetPayload(), restapi.GetUssAvailabilityOperationID)
	}
	return &getUssAvailabilityPayload{Uss: dssmodels.Manager(req.GetUss())}, nil
}

func encodeSetUssAvailability(request dssstore.OperationRequest) (*raftpb.Payload, error) {
	payload, ok := request.(*setUssAvailabilityPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.SetUssAvailabilityOperationID)
	}
	availability, err := payload.Availability.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to encode USS availability state for operation %q", restapi.SetUssAvailabilityOperationID)
	}
	return ProposalPayload(&scdpb.Payload{Payload: &scdpb.Payload_SetUssAvailabilityOperation{
		SetUssAvailabilityOperation: &scdpb.SetUssAvailabilityOperation{
			Uss:          payload.Uss.String(),
			Availability: availability,
			OldVersion:   payload.OldVersion.ToProto(),
		},
	}}), nil
}

func decodeSetUssAvailability(payload *raftpb.Payload) (dssstore.OperationRequest, error) {
	req := payload.GetScdPayload().GetSetUssAvailabilityOperation()
	if req == nil {
		return nil, stacktrace.NewError("unexpected payload %T for operation %q", payload.GetPayload(), restapi.SetUssAvailabilityOperationID)
	}
	decoded := &setUssAvailabilityPayload{Uss: dssmodels.Manager(req.GetUss())}
	if err := decoded.Availability.FromProto(req.GetAvailability()); err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode USS availability state for operation %q", restapi.SetUssAvailabilityOperationID)
	}
	if err := decoded.OldVersion.FromProto(req.GetOldVersion()); err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode USS old version for operation %q", restapi.SetUssAvailabilityOperationID)
	}
	return decoded, nil
}

type setUssAvailabilityPayload struct {
	Uss          dssmodels.Manager
	Availability scdmodels.UssAvailabilityState
	OldVersion   scdmodels.OVN
}

func (p *setUssAvailabilityPayload) OperationID() string {
	return restapi.SetUssAvailabilityOperationID
}

// NewSetUssAvailabilityPayload performs the request validation that can be done ahead of the
// transaction for a USS availability update request.
func NewSetUssAvailabilityPayload(ussID string, params *restapi.SetUssAvailabilityStatusParameters) (dssstore.OperationRequest, error) {
	availability, err := scdmodels.UssAvailabilityStateFromRest(params.Availability)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid availability state")
	}
	return &setUssAvailabilityPayload{
		Uss:          dssmodels.ManagerFromString(ussID),
		Availability: availability,
		OldVersion:   scdmodels.OVN(params.OldVersion),
	}, nil
}

func executeGetUssAvailability(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*getUssAvailabilityPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.GetUssAvailabilityOperationID)
	}

	id := payload.Uss

	// Get USS availability from Store
	ussa, err := repo.GetUssAvailability(ctx, id)
	if err != nil && !errors.Is(err, repos.ErrNotFound) {
		return nil, stacktrace.Propagate(err, "Could not get USS availability from repo")
	}
	if ussa == nil {
		// Return default availability status "Unknown"
		return &restapi.UssAvailabilityStatusResponse{
			Status: restapi.UssAvailabilityStatus{
				Availability: restapi.UssAvailabilityState_Unknown,
				Uss:          id.String()},
		}, nil
	}

	return &restapi.UssAvailabilityStatusResponse{
		Status:  *ussa.ToRest(),
		Version: ussa.Version.String(),
	}, nil
}

func executeSetUssAvailability(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*setUssAvailabilityPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, restapi.SetUssAvailabilityOperationID)
	}

	id := payload.Uss
	version := payload.OldVersion
	ussareq := &scdmodels.UssAvailabilityStatus{
		Uss:          id,
		Availability: payload.Availability,
	}

	old, err := repo.GetUssAvailability(ctx, id)
	if err != nil && !errors.Is(err, repos.ErrNotFound) {
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
