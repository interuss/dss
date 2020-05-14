package scd

import (
	"context"
	"fmt"
	"time"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	"github.com/interuss/dss/pkg/dss/auth"
	"github.com/interuss/dss/pkg/dss/geo"
  "github.com/interuss/dss/pkg/dss/models"

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

func DSSErrorOfAreaError(err error) error {
	switch err.(type) {
	case *geo.ErrAreaTooLarge:
		return dsserr.AreaTooLarge(err.Error())
	default:
		return dsserr.BadRequest(fmt.Sprintf("bad area: %s", err))
	}
}

func MakeSubscribersToNotify(subscriptions []*scdmodels.Subscription) []*scdpb.SubscriberToNotify {
  result := []*scdpb.SubscriberToNotify{}

  // Collect Subscriptions by URL
  subscriptions_by_url := map[string][]*scdpb.SubscriptionState{}
  for _, sub := range subscriptions {
    sub_state := &scdpb.SubscriptionState{
      SubscriptionId:    sub.ID.String(),
      NotificationIndex: sub.NotificationIndex,
    }
    if states, ok := subscriptions_by_url[sub.BaseURL]; ok {
      states = append(states, sub_state)
    } else {
      states = []*scdpb.SubscriptionState{sub_state}
    }
  }
  for url, states := range subscriptions_by_url {
    sub_proto := &scdpb.SubscriberToNotify{
      UssBaseUrl: url,
    }
    for _, state := range states {
      sub_proto.Subscriptions = append(sub_proto.Subscriptions, state)
    }
    result = append(result, sub_proto)
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
  op_proto, err := op.ToProto()
  if err != nil {
    return nil, dsserr.Internal(err.Error())
  }

  // Return response to client
  return &scdpb.ChangeOperationReferenceResponse{
    OperationReference: op_proto,
    Subscribers:        MakeSubscribersToNotify(subs),
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
  // Retrieve Operation ID
  idString := req.GetEntityuuid()
  if idString == "" {
    return nil, dsserr.BadRequest("missing Operation ID")
  }
  id := scdmodels.ID(idString)

  // Get Operation from Store
  sub, err := a.Store.GetOperation(ctx, id)
  if err != nil {
    return nil, err
  }
  if sub == nil {
    return nil, dsserr.Internal(fmt.Sprintf("GetOperation returned no Operation for ID: %s", id))
  }

  // Convert Operation to proto
  p, err := sub.ToProto()
  if err != nil {
    return nil, dsserr.Internal(err.Error())
  }

  // Return response to client
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

  var (
    params = req.GetParams()
  )

  // Parse extents
  for _, single_extents := range params.GetExtents() {
    // TODO: aggregate multiple Volume4's into a single extents
    extents, err := single_extents.ToCommon()
    if err != nil {
      return nil, dsserr.BadRequest(fmt.Sprintf("unable to parse extents: %s", err))
    }
  }

  // Process Subscription (creating implicit Subscription if necessary
  sub_id := scdmodels.ID(params.GetSubscriptionId())
  // TODO: Create implicit Subsription if requested

  // Parse the key
  key := []scdmodels.OVN{}
  if params.Key != nil {
    for _, ovn := range params.GetKey() {
      key = append(key, scdmodels.OVN{ovn}) //TODO: create OVN from string
    }
  }

  // Construct requested Operation model
  cells, err := extents.GetCells()
  if err != nil {
    return nil, DSSErrorOfAreaError(err)
  }
  op := &scdmodels.Operation{
    ID:            id,
    Owner:         owner,
    Version:       scdmodels.Version(params.OldVersion),

    StartTime:     extents.StartTime,
    EndTime:       extents.EndTime,
    AltitudeLower: extents.SpatialVolume.AltitudeLo,
    AltitudeUpper: extents.SpatialVolume.AltitudeHi,
    Cells:         cells,

    USSBaseURL:     params.UssBaseUrl,
    SubscriptionID: sub_id,
  }

  // Store Operation model
  op, subs, err := a.Store.UpsertOperation(ctx, op, key)
  if err != nil {
    return nil, err
  }
  if op == nil {
    return nil, dsserr.Internal(fmt.Sprintf("UpsertOperation returned no Operation for ID: %s", id))
  }

  // Convert Operation to proto
  p, err := op.ToProto()
  if err != nil {
    return nil, dsserr.Internal("could not convert Operation to proto")
  }

  // Return response to client
  return &scdpb.ChangeOperationReferenceResponse{
    OperationReference: p,
    Subscribers:        MakeSubscribersToNotify(subs),
  }, nil
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
	cells, err := extents.GetCells()
	if err != nil {
		return nil, DSSErrorOfAreaError(err)
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
	sub, err = a.Store.UpsertSubscription(ctx, sub)
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
	cells, err := vol4.GetCells()
	if err != nil {
		return nil, DSSErrorOfAreaError(err)
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
