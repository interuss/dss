package scd

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scderr "github.com/interuss/dss/pkg/scd/errors"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	"github.com/palantir/stacktrace"
	"google.golang.org/grpc/status"
)

// DeleteOperationReference deletes a single operation ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationReference(ctx context.Context, req *scdpb.DeleteOperationReferenceRequest) (*scdpb.ChangeOperationReferenceResponse, error) {
	// Retrieve Operation ID
	id, err := dssmodels.IDFromString(req.GetEntityuuid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.ChangeOperationReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Delete Operation in Store
		op, subs, err := r.DeleteOperation(ctx, id, owner)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to delete Operation from repo")
		}
		if op == nil {
			return stacktrace.NewError(fmt.Sprintf("DeleteOperation returned no Operation for ID: %s", id))
		}

		// Convert deleted Operation to proto
		opProto, err := op.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert Operation to proto")
		}

		// Return response to client
		response = &scdpb.ChangeOperationReferenceResponse{
			OperationReference: opProto,
			Subscribers:        makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// GetOperationReference returns a single operation ref for the given ID.
func (a *Server) GetOperationReference(ctx context.Context, req *scdpb.GetOperationReferenceRequest) (*scdpb.GetOperationReferenceResponse, error) {
	id, err := dssmodels.IDFromString(req.GetEntityuuid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var response *scdpb.GetOperationReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		op, err := r.GetOperation(ctx, id)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to get Operation from repo")
		}

		if op.Owner != owner {
			op.OVN = scdmodels.OVN("")
		}

		p, err := op.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert Operation to proto")
		}

		response = &scdpb.GetOperationReferenceResponse{
			OperationReference: p,
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// SearchOperationReferences queries existing operation refs in the given
// bounds.
func (a *Server) SearchOperationReferences(ctx context.Context, req *scdpb.SearchOperationReferencesRequest) (*scdpb.SearchOperationReferenceResponse, error) {
	// Retrieve the area of interest parameter
	aoi := req.GetParams().AreaOfInterest
	if aoi == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := dssmodels.Volume4DFromSCDProto(aoi)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Error parsing geometry")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	if aoi.TimeEnd != nil {
		endTime, _ := ptypes.Timestamp(aoi.TimeEnd.Value)
		if time.Now().After(endTime) {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "End time is in the past")
		}
	}

	var response *scdpb.SearchOperationReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		// Perform search query on Store
		ops, err := r.SearchOperations(ctx, vol4)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to search for Operations in repo")
		}

		// Create response for client
		response = &scdpb.SearchOperationReferenceResponse{}
		for _, op := range ops {
			p, err := op.ToProto()
			if err != nil {
				return stacktrace.Propagate(err, "Could not convert Operation model to proto")
			}
			if op.Owner != owner {
				p.Ovn = ""
			}
			response.OperationReferences = append(response.OperationReferences, p)
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}

// PutOperationReference creates a single operation ref.
func (a *Server) PutOperationReference(ctx context.Context, req *scdpb.PutOperationReferenceRequest) (*scdpb.ChangeOperationReferenceResponse, error) {
	id, err := dssmodels.IDFromString(req.GetEntityuuid())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}

	var (
		params  = req.GetParams()
		extents = make([]*dssmodels.Volume4D, len(params.GetExtents()))
	)

	if len(params.UssBaseUrl) == 0 {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required UssBaseUrl")
	}

	err = scdmodels.ValidateUSSBaseURL(params.UssBaseUrl)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL")
	}

	for idx, extent := range params.GetExtents() {
		cExtent, err := dssmodels.Volume4DFromSCDProto(extent)
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to parse extents")
		}
		extents[idx] = cExtent
	}
	uExtent, err := dssmodels.UnionVolumes4D(extents...)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to union extents")
	}

	if uExtent.StartTime == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing time_start from extents")
	}
	if uExtent.EndTime == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing time_end from extents")
	}

	cells, err := uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, dssErrorOfAreaError(err)
	}

	if uExtent.EndTime.Before(*uExtent.StartTime) {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "End time is past the start time")
	}

	if time.Now().After(*uExtent.EndTime) {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "End time is in the past")
	}

	if params.OldVersion == 0 && params.State != "Accepted" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid state for version 0")
	}

	subscriptionID, err := dssmodels.IDFromOptionalString(params.GetSubscriptionId())
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format for Subscription ID")
	}

	var response *scdpb.ChangeOperationReferenceResponse
	action := func(ctx context.Context, r repos.Repository) (err error) {
		if subscriptionID.Empty() {
			err := scdmodels.ValidateUSSBaseURL(params.GetNewSubscription().GetUssBaseUrl())
			if err != nil {
				return stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate USS base URL")
			}

			sub, err := r.UpsertSubscription(ctx, &scdmodels.Subscription{
				ID:         dssmodels.ID(uuid.New().String()),
				Owner:      owner,
				StartTime:  uExtent.StartTime,
				EndTime:    uExtent.EndTime,
				AltitudeLo: uExtent.SpatialVolume.AltitudeLo,
				AltitudeHi: uExtent.SpatialVolume.AltitudeHi,
				Cells:      cells,

				BaseURL:              params.GetNewSubscription().GetUssBaseUrl(),
				NotifyForOperations:  true,
				NotifyForConstraints: params.GetNewSubscription().GetNotifyForConstraints(),
				ImplicitSubscription: true,
			})
			if err != nil {
				return stacktrace.Propagate(err, "Failed to create implicit subscription")
			}
			subscriptionID = sub.ID
		} else {
			sub, err := r.GetSubscription(ctx, subscriptionID)
			if err != nil {
				return stacktrace.Propagate(err, "Unable to get Subscription")
			}
			if sub == nil {
				return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Specified Subscription does not exist")
			}
			updateSub := false
			if sub.StartTime != nil && sub.StartTime.After(*uExtent.StartTime) {
				if sub.ImplicitSubscription {
					sub.StartTime = uExtent.StartTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not begin until after the Operation starts")
				}
			}
			if sub.EndTime != nil && sub.EndTime.Before(*uExtent.EndTime) {
				if sub.ImplicitSubscription {
					sub.EndTime = uExtent.EndTime
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription ends before the Operation ends")
				}
			}
			if !sub.Cells.Contains(cells) {
				if sub.ImplicitSubscription {
					sub.Cells = s2.CellUnionFromUnion(sub.Cells, cells)
					updateSub = true
				} else {
					return stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not cover entire spatial area of the Operation")
				}
			}
			if updateSub {
				_, err := r.UpsertSubscription(ctx, sub)
				if err != nil {
					return stacktrace.Propagate(err, "Failed to update implicit Subscription")
				}
			}
		}

		key := []scdmodels.OVN{}
		for _, ovn := range params.GetKey() {
			key = append(key, scdmodels.OVN(ovn))
		}

		op, subs, err := r.UpsertOperation(ctx, &scdmodels.Operation{
			ID:      id,
			Owner:   owner,
			Version: scdmodels.Version(params.OldVersion),

			StartTime:     uExtent.StartTime,
			EndTime:       uExtent.EndTime,
			AltitudeLower: uExtent.SpatialVolume.AltitudeLo,
			AltitudeUpper: uExtent.SpatialVolume.AltitudeHi,
			Cells:         cells,

			USSBaseURL:     params.UssBaseUrl,
			SubscriptionID: subscriptionID,
			State:          scdmodels.OperationState(params.State),
		}, key)

		if err == scderr.MissingOVNsInternalError() {
			// The client is missing some OVNs; provide the pointers to the
			// information they need
			ops, err := r.SearchOperations(ctx, uExtent)
			if err != nil {
				return stacktrace.Propagate(err, "Could not search Operations in repo")
			}

			for _, op := range ops {
				if op.Owner != owner {
					op.OVN = ""
				}
			}

			p, err := scderr.MissingOVNsErrorResponse(ops)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to construct missing OVNs error message")
			}
			return status.ErrorProto(p)
		}

		if err != nil {
			if _, ok := status.FromError(err); ok {
				return err // No need to Propagate this error as Status errors are intended responses to client
			}
			return stacktrace.Propagate(err, "Failed to upsert operation in repo")
		}

		p, err := op.ToProto()
		if err != nil {
			return stacktrace.Propagate(err, "Could not convert Operation to proto")
		}

		response = &scdpb.ChangeOperationReferenceResponse{
			OperationReference: p,
			Subscribers:        makeSubscribersToNotify(subs),
		}

		return nil
	}

	err = a.Store.Transact(ctx, action)
	if err != nil {
		return nil, err // No need to Propagate this error as this is not a useful stacktrace line
	}

	return response, nil
}
