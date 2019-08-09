package dss

import (
	"context"
	"time"

	"github.com/steeling/InterUSS-Platform/pkg/dss/models"

	"github.com/golang/protobuf/ptypes"

	"github.com/steeling/InterUSS-Platform/pkg/dss/auth"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
	dsserr "github.com/steeling/InterUSS-Platform/pkg/errors"
)

var (
	WriteISAScope = "dss.write.identification_service_areas"
	ReadISAScope  = "dss.read.identification_service_areas"
)

func ptrToFloat32(f float32) *float32 {
	return &f
}

// Server implements dssproto.DiscoveryAndSynchronizationService.
type Server struct {
	Store Store
}

func (s *Server) AuthScopes() map[string][]string {
	return map[string][]string{
		"GetIdentificationServiceArea":     []string{ReadISAScope},
		"PutIdentificationServiceArea":     []string{WriteISAScope},
		"PatchIdentificationServiceArea":   []string{WriteISAScope},
		"DeleteIdentificationServiceArea":  []string{WriteISAScope},
		"PutSubscription":                  []string{ReadISAScope},
		"PatchSubscription":                []string{ReadISAScope},
		"DeleteSubscription":               []string{ReadISAScope},
		"SearchSubscriptions":              []string{ReadISAScope},
		"SearchIdentificationServiceAreas": []string{ReadISAScope},
	}
}

func (s *Server) PatchIdentificationServiceArea(ctx context.Context, req *dspb.PatchIdentificationServiceAreaRequest) (*dspb.PatchIdentificationServiceAreaResponse, error) {
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var (
		starts *time.Time
		ends   *time.Time
	)
	if startTime := req.GetExtents().GetStartTime(); startTime != nil {
		ts, err := ptypes.Timestamp(startTime)
		if err != nil {
			return nil, dsserr.BadRequest(err.Error())
		}
		starts = &ts
	}

	if endTime := req.GetExtents().GetEndTime(); endTime != nil {
		ts, err := ptypes.Timestamp(endTime)
		if err != nil {
			return nil, dsserr.BadRequest(err.Error())
		}
		ends = &ts
	}

	updated, err := models.VersionStringToTimestamp(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(err.Error())
	}

	isa := &models.IdentificationServiceArea{
		ID:        req.GetId(),
		Url:       req.GetUrl().GetValue(),
		Owner:     owner,
		Cells:     geo.Volume4DToCellIDs(req.GetExtents()),
		StartTime: starts,
		EndTime:   ends,
		UpdatedAt: &updated,
	}
	if wrapper := req.GetExtents().GetSpatialVolume().GetAltitudeHi(); wrapper != nil {
		isa.AltitudeHi = ptrToFloat32(wrapper.GetValue())
	}
	if wrapper := req.GetExtents().GetSpatialVolume().GetAltitudeLo(); wrapper != nil {
		isa.AltitudeLo = ptrToFloat32(wrapper.GetValue())
	}

	isa, subscribers, err := s.Store.UpdateISA(ctx, isa)
	if err != nil {
		return nil, err
	}

	pbISA, err := isa.ToProto()
	if err == nil {
		return nil, dsserr.Internal(err.Error())
	}

	pbSubscribers := []*dspb.SubscriberToNotify{}
	for _, subscriber := range subscribers {
		pbSubscribers = append(pbSubscribers, subscriber.ToNotifyProto())
	}

	return &dspb.PatchIdentificationServiceAreaResponse{
		ServiceArea: pbISA,
		Subscribers: pbSubscribers,
	}, nil
}

func (s *Server) PutIdentificationServiceArea(ctx context.Context, req *dspb.PutIdentificationServiceAreaRequest) (*dspb.PutIdentificationServiceAreaResponse, error) {
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	var (
		starts *time.Time
		ends   *time.Time
	)
	if startTime := req.GetExtents().GetStartTime(); startTime != nil {
		ts, err := ptypes.Timestamp(startTime)
		if err != nil {
			return nil, dsserr.BadRequest(err.Error())
		}
		starts = &ts
	}

	if endTime := req.GetExtents().GetEndTime(); endTime != nil {
		ts, err := ptypes.Timestamp(endTime)
		if err != nil {
			return nil, dsserr.BadRequest(err.Error())
		}
		ends = &ts
	}

	isa := &models.IdentificationServiceArea{
		ID:        req.GetId(),
		Url:       req.GetUrl(),
		Owner:     owner,
		Cells:     geo.Volume4DToCellIDs(req.GetExtents()),
		StartTime: starts,
		EndTime:   ends,
	}
	if wrapper := req.GetExtents().GetSpatialVolume().GetAltitudeHi(); wrapper != nil {
		isa.AltitudeHi = ptrToFloat32(wrapper.GetValue())
	}
	if wrapper := req.GetExtents().GetSpatialVolume().GetAltitudeLo(); wrapper != nil {
		isa.AltitudeLo = ptrToFloat32(wrapper.GetValue())
	}

	isa, subscribers, err := s.Store.InsertISA(ctx, isa)
	if err != nil {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, err
	}

	pbISA, err := isa.ToProto()
	if err == nil {
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

func (s *Server) DeleteIdentificationServiceArea(ctx context.Context, req *dspb.DeleteIdentificationServiceAreaRequest) (*dspb.DeleteIdentificationServiceAreaResponse, error) {
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	isa, subscribers, err := s.Store.DeleteISA(ctx, req.GetId(), owner, req.Version)
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

func (s *Server) DeleteSubscription(ctx context.Context, req *dspb.DeleteSubscriptionRequest) (*dspb.DeleteSubscriptionResponse, error) {
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	subscription, err := s.Store.DeleteSubscription(ctx, req.GetId(), owner, req.GetVersion())
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

func (s *Server) SearchIdentificationServiceAreas(ctx context.Context, req *dspb.SearchIdentificationServiceAreasRequest) (*dspb.SearchIdentificationServiceAreasResponse, error) {
	cu, err := geo.AreaToCellIDs(req.GetArea())
	if err != nil {
		return nil, dsserr.Internal(err.Error())
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

func (s *Server) SearchSubscriptions(ctx context.Context, req *dspb.SearchSubscriptionsRequest) (*dspb.SearchSubscriptionsResponse, error) {
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}

	cu, err := geo.AreaToCellIDs(req.GetArea())
	if err != nil {
		return nil, err
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

func (s *Server) GetSubscription(ctx context.Context, req *dspb.GetSubscriptionRequest) (*dspb.GetSubscriptionResponse, error) {
	subscription, err := s.Store.GetSubscription(ctx, req.GetId())
	if err != nil {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
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

func (s *Server) PatchSubscription(ctx context.Context, req *dspb.PatchSubscriptionRequest) (*dspb.PatchSubscriptionResponse, error) {
	return nil, nil
}

func (s *Server) PutSubscription(ctx context.Context, req *dspb.PutSubscriptionRequest) (*dspb.PutSubscriptionResponse, error) {
	return nil, nil
}
