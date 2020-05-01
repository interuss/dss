package utm

import (
	"context"
	"fmt"
	"time"

	"github.com/interuss/dss/pkg/api/v1/utmpb"
	"github.com/interuss/dss/pkg/dss/auth"
	//"github.com/interuss/dss/pkg/dss/geo"
	"github.com/interuss/dss/pkg/dss/utm/models"
	dsserr "github.com/interuss/dss/pkg/errors"

	"github.com/golang/geo/s2"
	//"github.com/golang/protobuf/ptypes"
)

// Server implements utmpb.DiscoveryAndSynchronizationService.
type Server struct {
	Store   Store
	Timeout time.Duration
}

func (a *Server) DeleteConstraintReference(ctx context.Context, req *utmpb.DeleteConstraintReferenceRequest) (*utmpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) DeleteOperationReference(ctx context.Context, req *utmpb.DeleteOperationReferenceRequest) (*utmpb.ChangeOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) DeleteSubscription(ctx context.Context, req *utmpb.DeleteSubscriptionRequest) (*utmpb.DeleteSubscriptionResponse, error) {
  // Retrieve Subscription ID
  id_string := req.GetSubscriptionid()
  if id_string == "" {
    return nil, dsserr.BadRequest("missing Subscription ID")
  }
  id := models.ID(id_string)

  // Retrieve ID of client making call
  owner, ok := auth.OwnerFromContext(ctx)
  if !ok {
    return nil, dsserr.PermissionDenied("missing owner from context")
  }

  // Delete Subscription in Store
  sub, err := a.Store.DeleteSubscription(ctx, id, owner)
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
	return &utmpb.DeleteSubscriptionResponse {
    Subscription: p,
  }, nil
}

func (a *Server) GetConstraintReference(ctx context.Context, req *utmpb.GetConstraintReferenceRequest) (*utmpb.GetConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) GetOperationReference(ctx context.Context, req *utmpb.GetOperationReferenceRequest) (*utmpb.GetOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) GetSubscription(ctx context.Context, req *utmpb.GetSubscriptionRequest) (*utmpb.GetSubscriptionResponse, error) {
  // Retrieve Subscription ID
  id_string := req.GetSubscriptionid()
  if id_string == "" {
    return nil, dsserr.BadRequest("missing Subscription ID")
  }
  id := models.ID(id_string)

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
	return &utmpb.GetSubscriptionResponse {
    Subscription: p,
  }, nil
}

func (a *Server) MakeDssReport(ctx context.Context, req *utmpb.MakeDssReportRequest) (*utmpb.ErrorReport, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) PutConstraintReference(ctx context.Context, req *utmpb.PutConstraintReferenceRequest) (*utmpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) PutOperationReference(ctx context.Context, req *utmpb.PutOperationReferenceRequest) (*utmpb.ChangeOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) PutSubscription(ctx context.Context, req *utmpb.PutSubscriptionRequest) (*utmpb.PutSubscriptionResponse, error) {
  // Retrieve Subscription ID
  id_string := req.GetSubscriptionid()
  if id_string == "" {
    return nil, dsserr.BadRequest("missing Subscription ID")
  }
  id := models.ID(id_string)

  // Retrieve ID of client making call
  owner, ok := auth.OwnerFromContext(ctx)
  if !ok {
    return nil, dsserr.PermissionDenied("missing owner from context")
  }

  // If this is an update, get the old version
  params := req.GetParams()
  var notification_index = 0
  var implicit_subscription = false
  var old_version = 0
  if params.OldVersion > 0 {
  //TODO: This needs to happen in a single transaction
    old_sub, err := a.Store.GetSubscription(ctx, id, owner)
    if err != nil {
      return nil, err //TODO: Change to 409 if Subscription didn't already exist
    }
    if old_sub.Version != int(params.OldVersion) {
      return nil, dsserr.Conflict("old_version does not match current version")
    }
    notification_index = old_sub.NotificationIndex
    implicit_subscription = old_sub.ImplicitSubscription
    old_version = old_sub.Version
  }

  // Construct Subscription model
  sub := &models.Subscription{
    ID:                id,
    Version:           old_version + 1,
    NotificationIndex: notification_index,
    Owner:             owner,

    BaseURL:              params.UssBaseUrl,
    NotifyForOperations:  params.NotifyForOperations,
    NotifyForConstraints: params.NotifyForConstraints,
    ImplicitSubscription: implicit_subscription,
	}
	//TODO: Set StartTime, EndTime, AltitudeHi, AltitudeLo, DependentOperations

	// Store Subscription model
  sub, err := a.Store.InsertSubscription(ctx, sub, owner) //TODO: This should be UpsertSubscription
  if err != nil {
    return nil, err
  }
  if sub == nil {
    return nil, dsserr.Internal(fmt.Sprintf("InsertSubscription returned no Subscription for ID: %s", id))
  }

  //TODO: Search relevant Operations and Constraints

  // Convert Subscription to proto
  p, err := sub.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}

  // Return response to client
	return &utmpb.PutSubscriptionResponse {
    Subscription: p,
  }, nil
}

func (a *Server) QueryConstraintReferences(ctx context.Context, req *utmpb.QueryConstraintReferencesRequest) (*utmpb.SearchConstraintReferencesResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) QuerySubscriptions(ctx context.Context, req *utmpb.QuerySubscriptionsRequest) (*utmpb.SearchSubscriptionsResponse, error) {
  //Retrieve the area of interest parameter
  aoi := req.GetParams().AreaOfInterest
  if aoi == nil {
    return nil, dsserr.BadRequest("missing area_of_interest")
  }

  // Retrieve ID of client making call
  owner, ok := auth.OwnerFromContext(ctx)
  if !ok {
    return nil, dsserr.PermissionDenied("missing owner from context")
  }

//   volume = geo.ToVolume4(aoi)
// 	cu, err := geo.AreaToCellIDs(volume)
// 	if err != nil {
// 		errMsg := fmt.Sprintf("bad area: %s", err)
// 		switch err.(type) {
// 		case *geo.ErrAreaTooLarge:
// 			return nil, dsserr.AreaTooLarge(errMsg)
// 		}
// 		return nil, dsserr.BadRequest(errMsg)
// 	}
  cells := s2.CellUnion{} //TODO: Compute cells correctly

  // Perform search query on Store
  subs, err := a.Store.SearchSubscriptions(ctx, cells, owner) //TODO: incorporate time bounds into query
  if err != nil {
    return nil, err
  }
  if subs == nil {
    return nil, dsserr.Internal("SearchSubscriptions returned nil subscriptions")
  }

  // Return response to client
	response := &utmpb.SearchSubscriptionsResponse {}
  for _, sub := range subs {
    p, err := sub.ToProto()
    if err != nil {
      return nil, dsserr.Internal("error converting Subscription model to proto")
    }
    response.Subscriptions = append(response.Subscriptions, p)
  }
  return response, nil
}

func (a *Server) SearchOperationReferences(ctx context.Context, req *utmpb.SearchOperationReferencesRequest) (*utmpb.SearchOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}
