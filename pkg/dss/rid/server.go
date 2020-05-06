package rid

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/interuss/dss/pkg/api/v1/dsspb"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	ridmodels "github.com/interuss/dss/pkg/dss/rid/models"

	"github.com/golang/protobuf/ptypes"

	"github.com/interuss/dss/pkg/dss/auth"
	"github.com/interuss/dss/pkg/dss/geo"
	dsserr "github.com/interuss/dss/pkg/errors"
)

var (
	writeISAScope = "dss.write.identification_service_areas"
	readISAScope  = "dss.read.identification_service_areas"
)

// Server implements dsspb.DiscoveryAndSynchronizationService.
type Server struct {
	Store   Store
	Timeout time.Duration
}

// AuthScopes returns a map of endpoint to required Oauth scope.
func (s *Server) AuthScopes() map[string][]string {
	return map[string][]string{
		"CreateIdentificationServiceArea":  {writeISAScope},
		"DeleteIdentificationServiceArea":  {writeISAScope},
		"GetIdentificationServiceArea":     {readISAScope},
		"SearchIdentificationServiceAreas": {readISAScope},
		"UpdateIdentificationServiceArea":  {writeISAScope},
		"CreateSubscription":               {writeISAScope},
		"DeleteSubscription":               {writeISAScope},
		"GetSubscription":                  {readISAScope},
		"SearchSubscriptions":              {readISAScope},
		"UpdateSubscription":               {writeISAScope},
		"ValidateOauth":                    {writeISAScope},

		// TODO: replace with correct scopes
		"DeleteConstraintReference": {readISAScope}, //{constraintManagementScope},
		"DeleteOperationReference":  {readISAScope}, //{strategicCoordinationScope},
		// TODO: De-duplicate operation names
		//"DeleteSubscription":               {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
		"GetConstraintReference": {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
		"GetOperationReference":  {readISAScope}, //{strategicCoordinationScope},
		//"GetSubscription":                  {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
		"MakeDssReport":          {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
		"PutConstraintReference": {readISAScope}, //{constraintManagementScope},
		"PutOperationReference":  {readISAScope}, //{strategicCoordinationScope},
		//"PutSubscription":                  {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
		"QueryConstraintReferences": {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope, constraintManagementScope},
		"QuerySubscriptions":        {readISAScope}, //{strategicCoordinationScope, constraintConsumptionScope},
		"SearchOperationReferences": {readISAScope}, //{strategicCoordinationScope},
	}
}

// ===== Server =====

// GetIdentificationServiceArea returns a single ISA for a given ID.
func (s *Server) GetIdentificationServiceArea(
	ctx context.Context, req *dsspb.GetIdentificationServiceAreaRequest) (
	*dsspb.GetIdentificationServiceAreaResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isa, err := s.Store.GetISA(ctx, dssmodels.ID(req.GetId()))
	if err == sql.ErrNoRows {
		return nil, dsserr.NotFound(req.GetId())
	}
	if err != nil {
		return nil, err
	}
	p, err := isa.ToProto()
	if err != nil {
		return nil, err
	}
	return &dsspb.GetIdentificationServiceAreaResponse{
		ServiceArea: p,
	}, nil
}

func (s *Server) createOrUpdateISA(
	ctx context.Context, id string, version *dssmodels.Version, extents *dsspb.Volume4D, flightsURL string) (
	*dsspb.PutIdentificationServiceAreaResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	if flightsURL == "" {
		return nil, dsserr.BadRequest("missing required flightsURL")
	}
	if extents == nil {
		return nil, dsserr.BadRequest("missing required extents")
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:      dssmodels.ID(id),
		URL:     flightsURL,
		Owner:   owner,
		Version: version,
	}

	if err := isa.SetExtents(extents); err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad extents: %s", err))
	}

	insertedISA, subscribers, err := s.Store.InsertISA(ctx, isa)
	if err != nil {
		return nil, err
	}

	pbISA, err := insertedISA.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}

	pbSubscribers := []*dsspb.SubscriberToNotify{}
	for _, subscriber := range subscribers {
		pbSubscribers = append(pbSubscribers, subscriber.ToNotifyProto())
	}

	return &dsspb.PutIdentificationServiceAreaResponse{
		ServiceArea: pbISA,
		Subscribers: pbSubscribers,
	}, nil
}

// CreateIdentificationServiceArea creates an ISA
func (s *Server) CreateIdentificationServiceArea(
	ctx context.Context, req *dsspb.CreateIdentificationServiceAreaRequest) (
	*dsspb.PutIdentificationServiceAreaResponse, error) {

	params := req.GetParams()
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.createOrUpdateISA(ctx, req.GetId(), nil, params.Extents, params.GetFlightsUrl())
}

// UpdateIdentificationServiceArea updates an existing ISA.
func (s *Server) UpdateIdentificationServiceArea(
	ctx context.Context, req *dsspb.UpdateIdentificationServiceAreaRequest) (
	*dsspb.PutIdentificationServiceAreaResponse, error) {

	params := req.GetParams()

	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	return s.createOrUpdateISA(ctx, req.GetId(), version, params.Extents, params.GetFlightsUrl())
}

// DeleteIdentificationServiceArea deletes an existing ISA.
func (s *Server) DeleteIdentificationServiceArea(
	ctx context.Context, req *dsspb.DeleteIdentificationServiceAreaRequest) (
	*dsspb.DeleteIdentificationServiceAreaResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isa, subscribers, err := s.Store.DeleteISA(ctx, dssmodels.ID(req.GetId()), owner, version)
	if err != nil {
		return nil, err
	}

	p, err := isa.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}
	sp := make([]*dsspb.SubscriberToNotify, len(subscribers))
	for i := range subscribers {
		sp[i] = subscribers[i].ToNotifyProto()
	}

	return &dsspb.DeleteIdentificationServiceAreaResponse{
		ServiceArea: p,
		Subscribers: sp,
	}, nil
}

// DeleteSubscription deletes an existing subscription.
func (s *Server) DeleteSubscription(
	ctx context.Context, req *dsspb.DeleteSubscriptionRequest) (
	*dsspb.DeleteSubscriptionResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscription, err := s.Store.DeleteSubscription(ctx, dssmodels.ID(req.GetId()), owner, version)
	if err != nil {
		return nil, err
	}
	p, err := subscription.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}
	return &dsspb.DeleteSubscriptionResponse{
		Subscription: p,
	}, nil
}

// SearchIdentificationServiceAreas queries for all ISAs in the bounds.
func (s *Server) SearchIdentificationServiceAreas(
	ctx context.Context, req *dsspb.SearchIdentificationServiceAreasRequest) (
	*dsspb.SearchIdentificationServiceAreasResponse, error) {

	cu, err := geo.AreaToCellIDs(req.GetArea())
	if err != nil {
		errMsg := fmt.Sprintf("bad area: %s", err)
		switch err.(type) {
		case *geo.ErrAreaTooLarge:
			return nil, dsserr.AreaTooLarge(errMsg)
		}
		return nil, dsserr.BadRequest(errMsg)
	}

	var (
		earliest *time.Time
		latest   *time.Time
	)

	if et := req.GetEarliestTime(); et != nil {
		if ts, err := ptypes.Timestamp(et); err == nil {
			earliest = &ts
		} else {
			return nil, dsserr.Internal(err.Error())
		}
	}

	if lt := req.GetLatestTime(); lt != nil {
		if ts, err := ptypes.Timestamp(lt); err == nil {
			latest = &ts
		} else {
			return nil, dsserr.Internal(err.Error())
		}
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isas, err := s.Store.SearchISAs(ctx, cu, earliest, latest)
	if err != nil {
		return nil, err
	}

	areas := make([]*dsspb.IdentificationServiceArea, len(isas))
	for i := range isas {
		a, err := isas[i].ToProto()
		if err != nil {
			return nil, err
		}
		areas[i] = a
	}

	return &dsspb.SearchIdentificationServiceAreasResponse{
		ServiceAreas: areas,
	}, nil
}

// SearchSubscriptions queries for existing subscriptions in the given bounds.
func (s *Server) SearchSubscriptions(
	ctx context.Context, req *dsspb.SearchSubscriptionsRequest) (
	*dsspb.SearchSubscriptionsResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	cu, err := geo.AreaToCellIDs(req.GetArea())
	if err != nil {
		errMsg := fmt.Sprintf("bad area: %s", err)
		switch err.(type) {
		case *geo.ErrAreaTooLarge:
			return nil, dsserr.AreaTooLarge(errMsg)
		}
		return nil, dsserr.BadRequest(errMsg)
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscriptions, err := s.Store.SearchSubscriptions(ctx, cu, owner)
	if err != nil {
		return nil, err
	}
	sp := make([]*dsspb.Subscription, len(subscriptions))
	for i := range subscriptions {
		sp[i], err = subscriptions[i].ToProto()
		if err != nil {
			return nil, err
		}
	}

	return &dsspb.SearchSubscriptionsResponse{
		Subscriptions: sp,
	}, nil
}

// GetSubscription gets a single subscription based on ID.
func (s *Server) GetSubscription(
	ctx context.Context, req *dsspb.GetSubscriptionRequest) (
	*dsspb.GetSubscriptionResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscription, err := s.Store.GetSubscription(ctx, dssmodels.ID(req.GetId()))
	if err == sql.ErrNoRows {
		return nil, dsserr.NotFound(req.GetId())
	}
	if err != nil {
		return nil, err
	}
	p, err := subscription.ToProto()
	if err != nil {
		return nil, err
	}
	return &dsspb.GetSubscriptionResponse{
		Subscription: p,
	}, nil
}

func (s *Server) createOrUpdateSubscription(
	ctx context.Context, id string, version *dssmodels.Version, callbacks *dsspb.SubscriptionCallbacks, extents *dsspb.Volume4D) (
	*dsspb.PutSubscriptionResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	if callbacks == nil {
		return nil, dsserr.BadRequest("missing required callbacks")
	}
	if extents == nil {
		return nil, dsserr.BadRequest("missing required extents")
	}

	sub := &ridmodels.Subscription{
		ID:      dssmodels.ID(id),
		Owner:   owner,
		URL:     callbacks.IdentificationServiceAreaUrl,
		Version: version,
	}

	if err := sub.SetExtents(extents); err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad extents: %s", err))
	}

	insertedSub, err := s.Store.InsertSubscription(ctx, sub)
	if err != nil {
		return nil, err
	}

	p, err := insertedSub.ToProto()
	if err != nil {
		return nil, err
	}

	// Find ISAs that were in this subscription's area.
	isas, err := s.Store.SearchISAs(ctx, sub.Cells, nil, nil)
	if err != nil {
		return nil, err
	}

	// Convert the ISAs to protos.
	isaProtos := make([]*dsspb.IdentificationServiceArea, len(isas))
	for i, isa := range isas {
		isaProtos[i], err = isa.ToProto()
		if err != nil {
			return nil, err
		}
	}

	return &dsspb.PutSubscriptionResponse{
		Subscription: p,
		ServiceAreas: isaProtos,
	}, nil
}

// CreateSubscription creates a single subscription.
func (s *Server) CreateSubscription(
	ctx context.Context, req *dsspb.CreateSubscriptionRequest) (
	*dsspb.PutSubscriptionResponse, error) {

	params := req.GetParams()
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.createOrUpdateSubscription(ctx, req.GetId(), nil, params.Callbacks, params.Extents)
}

// UpdateSubscription updates a single subscription.
func (s *Server) UpdateSubscription(
	ctx context.Context, req *dsspb.UpdateSubscriptionRequest) (
	*dsspb.PutSubscriptionResponse, error) {

	params := req.GetParams()

	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.createOrUpdateSubscription(ctx, req.GetId(), version, params.Callbacks, params.Extents)
}
