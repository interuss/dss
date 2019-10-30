package dss

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/interuss/dss/pkg/dss/models"

	"github.com/golang/protobuf/ptypes"

	"github.com/interuss/dss/pkg/dss/auth"
	"github.com/interuss/dss/pkg/dss/geo"
	dspb "github.com/interuss/dss/pkg/dssproto"
	dsserr "github.com/interuss/dss/pkg/errors"
)

var (
	WriteISAScope = "dss.write.identification_service_areas"
	ReadISAScope  = "dss.read.identification_service_areas"
)

// Server implements dssproto.DiscoveryAndSynchronizationService.
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

func (s *Server) GetV1DssIdentificationServiceAreasId(
	ctx context.Context, req *dspb.GetV1DssIdentificationServiceAreasIdRequest) (
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
	*dspb.PutIdentificationServiceAreaResponse, error) {

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

	isa := models.IdentificationServiceArea{
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

	pbSubscribers := []*dspb.SubscriberToNotify{}
	for _, subscriber := range subscribers {
		pbSubscribers = append(pbSubscribers, subscriber.ToNotifyProto())
	}

	return &dspb.PutIdentificationServiceAreaResponse{
		ServiceArea: pbISA,
		Subscribers: pbSubscribers,
	}, nil
}

func (s *Server) PutV1DssIdentificationServiceAreasId(
	ctx context.Context, req *dspb.PutV1DssIdentificationServiceAreasIdRequest) (
	*dspb.PutIdentificationServiceAreaResponse, error) {

	params := req.GetParams()
	return s.createOrUpdateISA(ctx, req.GetId(), nil, params.Extents, params.GetFlightsUrl())
}

func (s *Server) PutV1DssIdentificationServiceAreasIdVersion(
	ctx context.Context, req *dspb.PutV1DssIdentificationServiceAreasIdVersionRequest) (
	*dspb.PutIdentificationServiceAreaResponse, error) {

	params := req.GetParams()

	version, err := models.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}

	return s.createOrUpdateISA(ctx, req.GetId(), version, params.Extents, params.GetFlightsUrl())
}

func (s *Server) DeleteV1DssIdentificationServiceAreasIdVersion(
	ctx context.Context, req *dspb.DeleteV1DssIdentificationServiceAreasIdVersionRequest) (
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

func (s *Server) DeleteV1DssSubscriptionsIdVersion(
	ctx context.Context, req *dspb.DeleteV1DssSubscriptionsIdVersionRequest) (
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

func (s *Server) GetV1DssIdentificationServiceAreas(
	ctx context.Context, req *dspb.GetV1DssIdentificationServiceAreasRequest) (
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

func (s *Server) GetV1DssSubscriptions(
	ctx context.Context, req *dspb.GetV1DssSubscriptionsRequest) (
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

func (s *Server) GetV1DssSubscriptionsId(
	ctx context.Context, req *dspb.GetV1DssSubscriptionsIdRequest) (
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
	*dspb.PutSubscriptionResponse, error) {

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
	var (
		// if TimeEnd is not supplied default values to 0
		// later in the code it will defaulted to 1 day after TimeStart
		endTime   = time.Unix(extents.TimeEnd.GetSeconds(), 0)
		startTime = time.Unix(extents.TimeStart.GetSeconds(), 0)
		diff      = endTime.Sub(startTime)
	)
	if diff > models.MaxSubscriptionDuration {
		return nil, dsserr.BadRequest("subscription window exceeds 24 hours")
	}

	sub := models.Subscription{
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
	isaProtos := make([]*dspb.IdentificationServiceArea, len(isas))
	for i, isa := range isas {
		isaProtos[i], err = isa.ToProto()
		if err != nil {
			return nil, err
		}
	}

	return &dspb.PutSubscriptionResponse{
		Subscription: p,
		ServiceAreas: isaProtos,
	}, nil
}

func (s *Server) PutV1DssSubscriptionsId(
	ctx context.Context, req *dspb.PutV1DssSubscriptionsIdRequest) (
	*dspb.PutSubscriptionResponse, error) {

	params := req.GetParams()
	return s.createOrUpdateSubscription(ctx, req.GetId(), nil, params.Callbacks, params.Extents)
}

func (s *Server) PutV1DssSubscriptionsIdVersion(
	ctx context.Context, req *dspb.PutV1DssSubscriptionsIdVersionRequest) (
	*dspb.PutSubscriptionResponse, error) {

	params := req.GetParams()

	version, err := models.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}

	return s.createOrUpdateSubscription(ctx, req.GetId(), version, params.Callbacks, params.Extents)
}
