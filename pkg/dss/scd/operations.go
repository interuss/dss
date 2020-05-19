package scd

import (
	"context"
	"fmt"

	"github.com/gogo/status"
	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/dss/auth"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	scdmodels "github.com/interuss/dss/pkg/dss/scd/models"
	dsserr "github.com/interuss/dss/pkg/errors"
)

// DeleteOperationReference deletes a single operation ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationReference(ctx context.Context, req *scdpb.DeleteOperationReferenceRequest) (*scdpb.ChangeOperationReferenceResponse, error) {
	// Retrieve Operation ID
	idString := req.GetEntityuuid()
	if idString == "" {
		return nil, dsserr.BadRequest("missing Operation ID")
	}
	id := scdmodels.ID(idString)

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	// Delete Operation in Store
	op, subs, err := a.Store.DeleteOperation(ctx, id, owner)
	if err != nil {
		return nil, err
	}
	if op == nil {
		return nil, dsserr.Internal(fmt.Sprintf("DeleteOperation returned no Operation for ID: %s", id))
	}
	if subs == nil {
		return nil, dsserr.Internal(fmt.Sprintf("DeleteOperation returned nil Subscriptions for ID: %s", id))
	}

	// Convert deleted Operation to proto
	opProto, err := op.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}

	// Return response to client
	return &scdpb.ChangeOperationReferenceResponse{
		OperationReference: opProto,
		Subscribers:        makeSubscribersToNotify(subs),
	}, nil
}

// GetOperationReference returns a single operation ref for the given ID.
func (a *Server) GetOperationReference(ctx context.Context, req *scdpb.GetOperationReferenceRequest) (*scdpb.GetOperationReferenceResponse, error) {
	id := scdmodels.ID(req.GetEntityuuid())
	if id.Empty() {
		return nil, dsserr.BadRequest("missing Operation ID")
	}

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	sub, err := a.Store.GetOperation(ctx, id)
	if err != nil {
		return nil, err
	}

	if sub.Owner != owner {
		sub.OVN = scdmodels.OVN("")
	}

	p, err := sub.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}

	return &scdpb.GetOperationReferenceResponse{
		OperationReference: p,
	}, nil
}

// SearchOperationReferences queries existing operation refs in the given
// bounds.
func (a *Server) SearchOperationReferences(ctx context.Context, req *scdpb.SearchOperationReferencesRequest) (*scdpb.SearchOperationReferenceResponse, error) {
	// Retrieve the area of interest parameter
	aoi := req.GetParams().AreaOfInterest
	if aoi == nil {
		return nil, dsserr.BadRequest("missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := dssmodels.Volume4DFromSCDProto(aoi)
	if err != nil {
		return nil, dsserr.Internal("failed to convert to internal geometry model")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	// Perform search query on Store
	ops, err := a.Store.SearchOperations(ctx, vol4, owner)
	if err != nil {
		return nil, err
	}

	// Return response to client
	response := &scdpb.SearchOperationReferenceResponse{}
	for _, op := range ops {
		p, err := op.ToProto()
		if err != nil {
			return nil, dsserr.Internal("error converting Operation model to proto")
		}
		response.OperationReferences = append(response.OperationReferences, p)
	}
	return response, nil
}

// PutOperationReference creates a single operation ref.
func (a *Server) PutOperationReference(ctx context.Context, req *scdpb.PutOperationReferenceRequest) (*scdpb.ChangeOperationReferenceResponse, error) {
	id := scdmodels.ID(req.GetEntityuuid())
	if id.Empty() {
		return nil, dsserr.BadRequest("missing Operation ID")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var (
		params  = req.GetParams()
		extents = make([]*scdmodels.Volume4D, len(params.GetExtents()))
	)

	if len(params.UssBaseUrl) == 0 {
		return nil, dsserr.BadRequest("missing required UssBaseUrl")
	}

	for idx, extent := range params.GetExtents() {
		cExtent, err := dssmodels.Volume4DFromSCDProto(extent)
		if err != nil {
			return nil, dsserr.BadRequest(fmt.Sprintf("failed to parse extents: %s", err))
		}
		extents[idx] = cExtent
	}
	uExtent, err := dssmodels.UnionVolumes4D(extents...)
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("failed to union extents: %s", err))
	}

	if uExtent.StartTime == nil {
		return nil, dsserr.BadRequest("missing time_start from extents")
	}
	if uExtent.EndTime == nil {
		return nil, dsserr.BadRequest("missing time_end from extents")
	}

	cells, err := uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, dssErrorOfAreaError(err)
	}

	subscriptionID := scdmodels.ID(params.GetSubscriptionId())

	if subscriptionID.Empty() {
		if err := scdmodels.ValidateUSSBaseURL(
			params.GetNewSubscription().GetUssBaseUrl(),
		); err != nil {
			return nil, dsserr.BadRequest(err.Error())
		}
		// TODO(tvoss): Creation of the subscription and the operation is not
		// atomic. That is, if the creation of the operation fails, we need to
		// rollback this subscription, too. See
		// https://github.com/interuss/dss/issues/277 for tracking purposes.
		sub, _, err := a.putSubscription(ctx, &scdmodels.Subscription{
			ID:         scdmodels.ID(uuid.New().String()),
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
			return nil, dsserr.Internal(fmt.Sprintf("failed to create implicit subscription: %s", err))
		}
		subscriptionID = sub.ID
	}

	key := []scdmodels.OVN{}
	for _, ovn := range params.GetKey() {
		key = append(key, scdmodels.OVN(ovn))
	}

	op, subs, err := a.Store.UpsertOperation(ctx, &scdmodels.Operation{
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

	if err != nil {
		if _, ok := status.FromError(err); ok {
			return nil, err
		}
		return nil, dsserr.Internal(fmt.Sprintf("failed to upsert operation: %s", err))
	}

	p, err := op.ToProto()
	if err != nil {
		return nil, dsserr.Internal("could not convert Operation to proto")
	}

	return &scdpb.ChangeOperationReferenceResponse{
		OperationReference: p,
		Subscribers:        makeSubscribersToNotify(subs),
	}, nil
}
