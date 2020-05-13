package scd

import (
	"context"
	"fmt"
	"time"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/dss/auth"

	//"github.com/interuss/dss/pkg/dss/geo"

	scdmodels "github.com/interuss/dss/pkg/dss/scd/models"
	dsserr "github.com/interuss/dss/pkg/errors"
	//"github.com/golang/protobuf/ptypes"
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

// DeleteConstraintReference deletes a single constraint ref for a given ID at
// the specified version.
func (a *Server) DeleteConstraintReference(ctx context.Context, req *scdpb.DeleteConstraintReferenceRequest) (*scdpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// DeleteOperationReference deletes a single operation ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationReference(ctx context.Context, req *scdpb.DeleteOperationReferenceRequest) (*scdpb.ChangeOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
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
	return nil, dsserr.BadRequest("not yet implemented")
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
	return nil, dsserr.BadRequest("not yet implemented")
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

	// If this is an update, get the old version
	var (
		params = req.GetParams()
	)

	// Construct Subscription model
	sub := &scdmodels.Subscription{
		ID:      id,
		Owner:   owner,
		Version: scdmodels.Version(params.OldVersion),

		BaseURL:              params.UssBaseUrl,
		NotifyForOperations:  params.NotifyForOperations,
		NotifyForConstraints: params.NotifyForConstraints,
	}
	//TODO: Set StartTime, EndTime, AltitudeHi, AltitudeLo, DependentOperations

	// Store Subscription model
	sub, err := a.Store.UpsertSubscription(ctx, sub) //TODO: This should be UpsertSubscription
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, dsserr.Internal(fmt.Sprintf("UpsertSubscription returned no Subscription for ID: %s", id))
	}

	//TODO: Search relevant Operations and Constraints

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
	//Retrieve the area of interest parameter
	aoi := req.GetParams().AreaOfInterest
	if aoi == nil {
		return nil, dsserr.BadRequest("missing area_of_interest")
	}

	caoi, err := aoi.ToCommon()
	if err != nil {
		return nil, dsserr.Internal("failed to convert to internal geometry model")
	}

	// Retrieve ID of client making call
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	cells, err := caoi.SpatialVolume.Footprint.CalculateCovering()
	if err != nil {
		return nil, dsserr.Internal("failed to calculate covering for geometry")
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
	return nil, dsserr.BadRequest("not yet implemented")
}
