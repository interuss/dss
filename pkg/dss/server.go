package dss

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/steeling/InterUSS-Platform/pkg/dss/models"

	"github.com/golang/protobuf/ptypes"

	"github.com/steeling/InterUSS-Platform/pkg/dss/auth"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dss_v1"
	dsserr "github.com/steeling/InterUSS-Platform/pkg/errors"
)

var (
	WriteISAScope = "dss.write.identification_service_areas"
	ReadISAScope  = "dss.read.identification_service_areas"
)

// Server implements dss_v1.DiscoveryAndSynchronizationService.
type Server struct {
	Store Store
}

func (s *Server) AuthScopes() map[string][]string {
	return map[string][]string{
		"GetIdentificationServiceArea":     []string{ReadISAScope},
		"PutIdentificationServiceArea":     []string{WriteISAScope},
		"DeleteIdentificationServiceArea":  []string{WriteISAScope},
		"PutSubscription":                  []string{ReadISAScope},
		"DeleteSubscription":               []string{ReadISAScope},
		"SearchSubscriptions":              []string{ReadISAScope},
		"SearchIdentificationServiceAreas": []string{ReadISAScope},
	}
}

func (s *Server) GetIdentificationServiceArea(
	ctx context.Context, req *dspb.GetIdentificationServiceAreaRequest) (
	*dspb.GetIdentificationServiceAreaResponse, error) {

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
	return &dspb.GetIdentificationServiceAreaResponse{
		ServiceArea: p,
	}, nil
}

func (s *Server) createOrUpdateISA(
	ctx context.Context, id string, version *models.Version, extents *dspb.Volume4D, flights_url string) (
	*dspb.IdentificationServiceArea, []*dspb.SubscriberToNotify, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, nil, dsserr.PermissionDenied("missing owner from context")
	}
	if flights_url == "" {
		return nil, nil, dsserr.BadRequest("missing required flights_url")
	}
	if extents == nil {
		return nil, nil, dsserr.BadRequest("missing required extents")
	}

	isa := models.IdentificationServiceArea{
		ID:      models.ID(id),
		Url:     flights_url,
		Owner:   owner,
		Version: version,
	}

	if err := isa.SetExtents(extents); err != nil {
		return nil, nil, dsserr.BadRequest(fmt.Sprintf("bad extents: %s", err))
	}

	insertedISA, subscribers, err := s.Store.InsertISA(ctx, isa)
	if err != nil {
		return nil, nil, err
	}

	pbISA, err := insertedISA.ToProto()
	if err != nil {
		return nil, nil, dsserr.Internal(err.Error())
	}

	pbSubscribers := []*dspb.SubscriberToNotify{}
	for _, subscriber := range subscribers {
		pbSubscribers = append(pbSubscribers, subscriber.ToNotifyProto())
	}

	return pbISA, pbSubscribers, nil
}

func (s *Server) CreateIdentificationServiceArea(
	ctx context.Context, req *dspb.CreateIdentificationServiceAreaRequest) (
	*dspb.CreateIdentificationServiceAreaResponse, error) {

	params := req.GetParams()
	isa, subs, err := s.createOrUpdateISA(ctx, req.GetId(), nil, params.Extents, params.GetFlightsUrl())
	return &dspb.CreateIdentificationServiceAreaResponse{ServiceArea: isa, Subscribers: subs}, err
}

func (s *Server) UpdateIdentificationServiceArea(
	ctx context.Context, req *dspb.UpdateIdentificationServiceAreaRequest) (
	*dspb.UpdateIdentificationServiceAreaResponse, error) {

	params := req.GetParams()

	version, err := models.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}

	isa, subs, err := s.createOrUpdateISA(ctx, req.GetId(), version, params.Extents, params.GetFlightsUrl())
	return &dspb.UpdateIdentificationServiceAreaResponse{ServiceArea: isa, Subscribers: subs}, err
}

func (s *Server) DeleteIdentificationServiceArea(
	ctx context.Context, req *dspb.DeleteIdentificationServiceAreaRequest) (
	*dspb.DeleteIdentificationServiceAreaResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	version, err := models.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	isa, subscribers, err := s.Store.DeleteISA(ctx, models.ID(req.GetId()), owner, version)
	if err != nil {
		return nil, err
	}

	p, err := isa.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}
	sp := make([]*dspb.SubscriberToNotify, len(subscribers))
	for i, _ := range subscribers {
		sp[i] = subscribers[i].ToNotifyProto()
	}

	return &dspb.DeleteIdentificationServiceAreaResponse{
		ServiceArea: p,
		Subscribers: sp,
	}, nil
}

func (s *Server) DeleteSubscription(
	ctx context.Context, req *dspb.DeleteSubscriptionRequest) (
	*dspb.DeleteSubscriptionResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	version, err := models.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	subscription, err := s.Store.DeleteSubscription(ctx, models.ID(req.GetId()), owner, version)
	if err != nil {
		return nil, err
	}
	p, err := subscription.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
	}
	return &dspb.DeleteSubscriptionResponse{
		Subscription: p,
	}, nil
}

func (s *Server) SearchIdentificationServiceAreas(
	ctx context.Context, req *dspb.SearchIdentificationServiceAreasRequest) (
	*dspb.SearchIdentificationServiceAreasResponse, error) {

	cu, err := geo.AreaToCellIDs(req.GetArea())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad area: %s", err))
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

	isas, err := s.Store.SearchISAs(ctx, cu, earliest, latest)
	if err != nil {
		return nil, err
	}

	areas := make([]*dspb.IdentificationServiceArea, len(isas))
	for i := range isas {
		a, err := isas[i].ToProto()
		if err != nil {
			return nil, err
		}
		areas[i] = a
	}

	return &dspb.SearchIdentificationServiceAreasResponse{
		ServiceAreas: areas,
	}, nil
}

func (s *Server) SearchSubscriptions(
	ctx context.Context, req *dspb.SearchSubscriptionsRequest) (
	*dspb.SearchSubscriptionsResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	cu, err := geo.AreaToCellIDs(req.GetArea())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad area: %s", err))
	}

	subscriptions, err := s.Store.SearchSubscriptions(ctx, cu, owner)
	if err != nil {
		return nil, err
	}
	sp := make([]*dspb.Subscription, len(subscriptions))
	for i, _ := range subscriptions {
		sp[i], err = subscriptions[i].ToProto()
		if err != nil {
			return nil, err
		}
	}

	return &dspb.SearchSubscriptionsResponse{
		Subscriptions: sp,
	}, nil
}

func (s *Server) GetSubscription(
	ctx context.Context, req *dspb.GetSubscriptionRequest) (
	*dspb.GetSubscriptionResponse, error) {

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
	return &dspb.GetSubscriptionResponse{
		Subscription: p,
	}, nil
}

func (s *Server) createOrUpdateSubscription(
	ctx context.Context, id string, version *models.Version, callbacks *dspb.SubscriptionCallbacks, extents *dspb.Volume4D) (
	*dspb.Subscription, []*dspb.IdentificationServiceArea, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, nil, dsserr.PermissionDenied("missing owner from context")
	}
	if callbacks == nil {
		return nil, nil, dsserr.BadRequest("missing required callbacks")
	}
	if extents == nil {
		return nil, nil, dsserr.BadRequest("missing required extents")
	}

	sub := models.Subscription{
		ID:      models.ID(id),
		Owner:   owner,
		Url:     callbacks.IdentificationServiceAreaUrl,
		Version: version,
	}

	if err := sub.SetExtents(extents); err != nil {
		return nil, nil, dsserr.BadRequest(fmt.Sprintf("bad extents: %s", err))
	}

	insertedSub, err := s.Store.InsertSubscription(ctx, sub)
	if err != nil {
		return nil, nil, err
	}

	p, err := insertedSub.ToProto()
	if err != nil {
		return nil, nil, err
	}

	// Find ISAs that were in this subscription's area.
	isas, err := s.Store.SearchISAs(ctx, sub.Cells, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	// Convert the ISAs to protos.
	isaProtos := make([]*dspb.IdentificationServiceArea, len(isas))
	for i, isa := range isas {
		isaProtos[i], err = isa.ToProto()
		if err != nil {
			return nil, nil, err
		}
	}

	return p, isaProtos, nil
}

func (s *Server) CreateSubscription(
	ctx context.Context, req *dspb.CreateSubscriptionRequest) (
	*dspb.CreateSubscriptionResponse, error) {

	params := req.GetParams()
	sub, isas, err := s.createOrUpdateSubscription(ctx, req.GetId(), nil, params.Callbacks, params.Extents)
	return &dspb.CreateSubscriptionResponse{Subscription: sub, ServiceAreas: isas}, err
}

func (s *Server) UpdateSubscription(
	ctx context.Context, req *dspb.UpdateSubscriptionsRequest) (
	*dspb.UpdateSubscriptionResponse, error) {

	params := req.GetParams()

	version, err := models.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}

	sub, isas, err := s.createOrUpdateSubscription(ctx, req.GetId(), version, params.Callbacks, params.Extents)
	return &dspb.UpdateSubscriptionResponse{Subscription: sub, ServiceAreas: isas}, err
}
