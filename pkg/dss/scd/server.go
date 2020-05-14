package scd

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/google/uuid"
	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/dss/auth"
	"github.com/interuss/dss/pkg/dss/geo"
	"github.com/interuss/dss/pkg/dss/models"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	scdmodels "github.com/interuss/dss/pkg/dss/scd/models"
	dsserr "github.com/interuss/dss/pkg/errors"
)

// Server implements scdpb.DiscoveryAndSynchronizationService.
type Server struct {
	Store   Store
	Timeout time.Duration
}

// AuthScopes returns a map of endpoint to required Oauth scope.
func (a *Server) AuthScopes() map[auth.Operation][]auth.Scope {
	// TODO: replace with correct scopes
	//"DeleteConstraintReference": {readISAScope}, //{constraintManagementScope},
	//"DeleteOperationReference":  {readISAScope}, //{strategicCoordinationScope},
	// TODO: De-duplicate operation names
	//"DeleteSubscription":               {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
	//"GetConstraintReference": {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
	//"GetOperationReference":  {readISAScope}, //{strategicCoordinationScope},
	//"GetSubscription":                  {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
	//"MakeDssReport":          {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
	//"PutConstraintReference": {readISAScope}, //{constraintManagementScope},
	//"PutOperationReference":  {readISAScope}, //{strategicCoordinationScope},
	//"PutSubscription":                  {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
	//"QueryConstraintReferences": {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
	//"QuerySubscriptions":        {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
	//"SearchOperationReferences": {readISAScope}, //{strategicCoordinationScope},
	return nil
}

func dssErrorOfAreaError(err error) error {
	switch err.(type) {
	case *geo.ErrAreaTooLarge:
		return dsserr.AreaTooLarge(err.Error())
	default:
		return dsserr.BadRequest(fmt.Sprintf("bad area: %s", err))
	}
}

func makeSubscribersToNotify(subscriptions []*scdmodels.Subscription) []*scdpb.SubscriberToNotify {
	result := []*scdpb.SubscriberToNotify{}

	subscriptionsByURL := map[string][]*scdpb.SubscriptionState{}
	for _, sub := range subscriptions {
		subState := &scdpb.SubscriptionState{
			SubscriptionId:    sub.ID.String(),
			NotificationIndex: int32(sub.NotificationIndex),
		}
		subscriptionsByURL[sub.BaseURL] = append(subscriptionsByURL[sub.BaseURL], subState)
	}
	for url, states := range subscriptionsByURL {
		result = append(result, &scdpb.SubscriberToNotify{
			UssBaseUrl:    url,
			Subscriptions: states,
		})
	}

	return result
}

// DeleteConstraintReference deletes a single constraint ref for a given ID at
// the specified version.
func (a *Server) DeleteConstraintReference(ctx context.Context, req *scdpb.DeleteConstraintReferenceRequest) (*scdpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

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

// DeleteSubscription deletes a single subscription for a given ID at the
// specified version.
func (a *Server) DeleteSubscription(ctx context.Context, req *scdpb.DeleteSubscriptionRequest) (*scdpb.DeleteSubscriptionResponse, error) {
	// Retrieve Subscription ID
	idString := req.GetSubscriptionid()
	if idString == "" {
		return nil, dsserr.BadRequest("missing Subscription ID")
	}
	id := scdmodels.ID(idString)

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	// Delete Subscription in Store
	sub, err := a.Store.DeleteSubscription(ctx, id, owner, scdmodels.Version(0))
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, dsserr.Internal(fmt.Sprintf("DeleteSubscription returned no Subscription for ID: %s", id))
	}

	// Convert deleted Subscription to proto
	p, err := sub.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}

	// Return response to client
	return &scdpb.DeleteSubscriptionResponse{
		Subscription: p,
	}, nil
}

// GetConstraintReference returns a single constraint ref for the given ID.
func (a *Server) GetConstraintReference(ctx context.Context, req *scdpb.GetConstraintReferenceRequest) (*scdpb.GetConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
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

// GetSubscription returns a single subscription for the given ID.
func (a *Server) GetSubscription(ctx context.Context, req *scdpb.GetSubscriptionRequest) (*scdpb.GetSubscriptionResponse, error) {
	// Retrieve Subscription ID
	idString := req.GetSubscriptionid()
	if idString == "" {
		return nil, dsserr.BadRequest("missing Subscription ID")
	}
	id := scdmodels.ID(idString)

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	// Get Subscription from Store
	sub, err := a.Store.GetSubscription(ctx, id, owner)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, dsserr.Internal(fmt.Sprintf("GetSubscription returned no Subscription for ID: %s", id))
	}

	// Convert Subscription to proto
	p, err := sub.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}

	// Return response to client
	return &scdpb.GetSubscriptionResponse{
		Subscription: p,
	}, nil
}

// MakeDssReport creates an error report about a DSS.
func (a *Server) MakeDssReport(ctx context.Context, req *scdpb.MakeDssReportRequest) (*scdpb.ErrorReport, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// PutConstraintReference creates a single contraint ref.
func (a *Server) PutConstraintReference(ctx context.Context, req *scdpb.PutConstraintReferenceRequest) (*scdpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
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
		extents = make([]*dssmodels.Volume4D, len(params.GetExtents()))
	)

	if len(params.UssBaseUrl) == 0 {
		return nil, dsserr.BadRequest("missing required UssBaseUrl")
	}

	for idx, extent := range params.GetExtents() {
		cExtent, err := extent.ToCommon()
		if err != nil {
			return nil, dsserr.BadRequest(fmt.Sprintf("failed to parse extents: %s", err))
		}
		extents[idx] = cExtent
	}
	uExtent, err := models.UnionVolumes4D(extents...)
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("failed to union extents: %s", err))
	}

	cells, err := uExtent.CalculateSpatialCovering()
	if err != nil {
		return nil, dssErrorOfAreaError(err)
	}

	subscriptionID := scdmodels.ID(params.GetSubscriptionId())

	if subscriptionID.Empty() {
		sub, err := a.putSubscription(ctx, &scdmodels.Subscription{
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
	}, key)
	if err != nil {
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

// PutSubscription creates a single subscription.
func (a *Server) putSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	// Store Subscription model
	return a.Store.UpsertSubscription(ctx, sub)
}

// PutSubscription creates a single subscription.
func (a *Server) PutSubscription(ctx context.Context, req *scdpb.PutSubscriptionRequest) (*scdpb.PutSubscriptionResponse, error) {
	// Retrieve Subscription ID
	idString := req.GetSubscriptionid()
	if idString == "" {
		return nil, dsserr.BadRequest("missing Subscription ID")
	}
	id := scdmodels.ID(idString)

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var (
		params = req.GetParams()
	)

	// Parse extents
	extents, err := params.GetExtents().ToCommon()
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("unable to parse extents: %s", err))
	}

	// Construct requested Subscription model
	cells, err := extents.CalculateSpatialCovering()
	switch err {
	case nil, dssmodels.ErrMissingSpatialVolume, dssmodels.ErrMissingFootprint:
		// All good, let's go ahead.
	default:
		return nil, dssErrorOfAreaError(err)
	}

	sub := &scdmodels.Subscription{
		ID:      id,
		Owner:   owner,
		Version: scdmodels.Version(params.OldVersion),

		StartTime:  extents.StartTime,
		EndTime:    extents.EndTime,
		AltitudeLo: extents.SpatialVolume.AltitudeLo,
		AltitudeHi: extents.SpatialVolume.AltitudeHi,
		Cells:      cells,

		BaseURL:              params.UssBaseUrl,
		NotifyForOperations:  params.NotifyForOperations,
		NotifyForConstraints: params.NotifyForConstraints,
	}

	// Store Subscription model
	sub, err = a.putSubscription(ctx, sub)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, dsserr.Internal(fmt.Sprintf("UpsertSubscription returned no Subscription for ID: %s", id))
	}

	// Convert Subscription to proto
	p, err := sub.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}

	// Return response to client
	return &scdpb.PutSubscriptionResponse{
		Subscription: p,
	}, nil
}

// QueryConstraintReferences queries existing contraint refs in the given
// bounds.
func (a *Server) QueryConstraintReferences(ctx context.Context, req *scdpb.QueryConstraintReferencesRequest) (*scdpb.SearchConstraintReferencesResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// QuerySubscriptions queries existing subscriptions in the given bounds.
func (a *Server) QuerySubscriptions(ctx context.Context, req *scdpb.QuerySubscriptionsRequest) (*scdpb.SearchSubscriptionsResponse, error) {
	// Retrieve the area of interest parameter
	aoi := req.GetParams().AreaOfInterest
	if aoi == nil {
		return nil, dsserr.BadRequest("missing area_of_interest")
	}

	// Parse area of interest to common Volume4D
	vol4, err := aoi.ToCommon()
	if err != nil {
		return nil, dsserr.Internal("failed to convert to internal geometry model")
	}

	// Extract S2 cells from area of interest
	cells, err := vol4.CalculateSpatialCovering()
	if err != nil {
		return nil, dssErrorOfAreaError(err)
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	// Perform search query on Store
	subs, err := a.Store.SearchSubscriptions(ctx, cells, owner) //TODO: incorporate time bounds into query
	if err != nil {
		return nil, err
	}
	if subs == nil {
		return nil, dsserr.Internal("SearchSubscriptions returned nil subscriptions")
	}

	// Return response to client
	response := &scdpb.SearchSubscriptionsResponse{}
	for _, sub := range subs {
		p, err := sub.ToProto()
		if err != nil {
			return nil, dsserr.Internal("error converting Subscription model to proto")
		}
		response.Subscriptions = append(response.Subscriptions, p)
	}
	return response, nil
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
	vol4, err := aoi.ToCommon()
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
	if ops == nil {
		return nil, dsserr.Internal("SearchOperations returned nil operations")
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
