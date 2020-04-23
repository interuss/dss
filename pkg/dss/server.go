package dss

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/interuss/dss/pkg/api/v1/auxpb"
	"github.com/interuss/dss/pkg/api/v1/dsspb"
	"github.com/interuss/dss/pkg/dss/models"

	"github.com/golang/protobuf/ptypes"

	"github.com/interuss/dss/pkg/dss/auth"
	"github.com/interuss/dss/pkg/dss/geo"
	dsserr "github.com/interuss/dss/pkg/errors"
)

var (
	WriteISAScope = "dss.write.identification_service_areas"
	ReadISAScope  = "dss.read.identification_service_areas"
)

// Server implements dsspb.DiscoveryAndSynchronizationService.
type Server struct {
	Store   Store
	Timeout time.Duration
}

// AuxServer implements auxpb.DSSAuxService.
type AuxServer struct{}

func (s *Server) AuthScopes() map[string][]string {
	return map[string][]string{
		"CreateIdentificationServiceArea":  {WriteISAScope},
		"DeleteIdentificationServiceArea":  {WriteISAScope},
		"GetIdentificationServiceArea":     {ReadISAScope},
		"SearchIdentificationServiceAreas": {ReadISAScope},
		"UpdateIdentificationServiceArea":  {WriteISAScope},
		"CreateSubscription":               {WriteISAScope},
		"DeleteSubscription":               {WriteISAScope},
		"GetSubscription":                  {ReadISAScope},
		"SearchSubscriptions":              {ReadISAScope},
		"UpdateSubscription":               {WriteISAScope},
		"ValidateOauth":                    {WriteISAScope},
	}
}

// Validate will exercise validating the Oauth token
func (a *AuxServer) ValidateOauth(ctx context.Context, req *auxpb.ValidateOauthRequest) (*auxpb.ValidateOauthResponse, error) {
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	if req.Owner != "" && req.Owner != owner.String() {
		return nil, dsserr.PermissionDenied(fmt.Sprintf("owner mismatch, required: %s, but oauth token has %s", req.Owner, owner))
	}
	return &auxpb.ValidateOauthResponse{}, nil
}

func (s *Server) GetIdentificationServiceArea(
	ctx context.Context, req *dsspb.GetIdentificationServiceAreaRequest) (
	*dsspb.GetIdentificationServiceAreaResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isa, err := s.Store.GetISA(ctx, models.ID(req.GetId()))
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
	ctx context.Context, id string, version *models.Version, extents *dsspb.Volume4D, flights_url string) (
	*dsspb.PutIdentificationServiceAreaResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	if flights_url == "" {
		return nil, dsserr.BadRequest("missing required flights_url")
	}
	if extents == nil {
		return nil, dsserr.BadRequest("missing required extents")
	}

	isa := &models.IdentificationServiceArea{
		ID:      models.ID(id),
		Url:     flights_url,
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

func (s *Server) CreateIdentificationServiceArea(
	ctx context.Context, req *dsspb.CreateIdentificationServiceAreaRequest) (
	*dsspb.PutIdentificationServiceAreaResponse, error) {

	params := req.GetParams()
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.createOrUpdateISA(ctx, req.GetId(), nil, params.Extents, params.GetFlightsUrl())
}

func (s *Server) UpdateIdentificationServiceArea(
	ctx context.Context, req *dsspb.UpdateIdentificationServiceAreaRequest) (
	*dsspb.PutIdentificationServiceAreaResponse, error) {

	params := req.GetParams()

	version, err := models.VersionFromString(req.GetVersion(), models.EmptyVersionPolicyRequireNonEmpty)
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	return s.createOrUpdateISA(ctx, req.GetId(), version, params.Extents, params.GetFlightsUrl())
}

func (s *Server) DeleteIdentificationServiceArea(
	ctx context.Context, req *dsspb.DeleteIdentificationServiceAreaRequest) (
	*dsspb.DeleteIdentificationServiceAreaResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	version, err := models.VersionFromString(req.GetVersion(), models.EmptyVersionPolicyRequireNonEmpty)
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isa, subscribers, err := s.Store.DeleteISA(ctx, models.ID(req.GetId()), owner, version)
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

func (s *Server) DeleteSubscription(
	ctx context.Context, req *dsspb.DeleteSubscriptionRequest) (
	*dsspb.DeleteSubscriptionResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	version, err := models.VersionFromString(req.GetVersion(), models.EmptyVersionPolicyRequireNonEmpty)
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscription, err := s.Store.DeleteSubscription(ctx, models.ID(req.GetId()), owner, version)
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

func (s *Server) GetSubscription(
	ctx context.Context, req *dsspb.GetSubscriptionRequest) (
	*dsspb.GetSubscriptionResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	subscription, err := s.Store.GetSubscription(ctx, models.ID(req.GetId()))
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
	ctx context.Context, id string, version *models.Version, callbacks *dsspb.SubscriptionCallbacks, extents *dsspb.Volume4D) (
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

	sub := &models.Subscription{
		ID:      models.ID(id),
		Owner:   owner,
		Url:     callbacks.IdentificationServiceAreaUrl,
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

func (s *Server) CreateSubscription(
	ctx context.Context, req *dsspb.CreateSubscriptionRequest) (
	*dsspb.PutSubscriptionResponse, error) {

	params := req.GetParams()
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.createOrUpdateSubscription(ctx, req.GetId(), nil, params.Callbacks, params.Extents)
}

func (s *Server) UpdateSubscription(
	ctx context.Context, req *dsspb.UpdateSubscriptionRequest) (
	*dsspb.PutSubscriptionResponse, error) {

	params := req.GetParams()

	version, err := models.VersionFromString(req.GetVersion(), models.EmptyVersionPolicyRequireNonEmpty)
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	return s.createOrUpdateSubscription(ctx, req.GetId(), version, params.Callbacks, params.Extents)
}
