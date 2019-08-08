package dss

import (
	"context"
	"errors"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/steeling/InterUSS-Platform/pkg/dss/auth"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
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
		"PatchIdentificationServiceArea":   []string{WriteISAScope},
		"DeleteIdentificationServiceArea":  []string{WriteISAScope},
		"PutSubscription":                  []string{ReadISAScope},
		"PatchSubscription":                []string{ReadISAScope},
		"DeleteSubscription":               []string{ReadISAScope},
		"SearchSubscriptions":              []string{ReadISAScope},
		"SearchIdentificationServiceAreas": []string{ReadISAScope},
	}
}

func (s *Server) DeleteIdentificationServiceArea(ctx context.Context, req *dspb.DeleteIdentificationServiceAreaRequest) (*dspb.DeleteIdentificationServiceAreaResponse, error) {
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, errors.New("missing owner from context")
	}

	isa, subscribers, err := s.Store.DeleteISA(ctx, req.GetId(), owner, req.Version)
	if err != nil {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, err
	}

	p, err := isa.ToProto()
	if err != nil {
		return nil, err
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
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, errors.New("missing owner from context")
	}
	subscription, err := s.Store.DeleteSubscription(ctx, req.GetId(), owner, req.GetVersion())
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
	return &dspb.DeleteSubscriptionResponse{
		Subscription: p,
	}, nil
}

func (s *Server) SearchIdentificationServiceAreas(ctx context.Context, req *dspb.SearchIdentificationServiceAreasRequest) (*dspb.SearchIdentificationServiceAreasResponse, error) {
	cu, err := geo.AreaToCellIDs(req.GetArea())
	if err != nil {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, err
	}

	var (
		earliest *time.Time
		latest   *time.Time
	)

	if et := req.GetEarliestTime(); et != nil {
		if ts, err := ptypes.Timestamp(et); err == nil {
			earliest = &ts
		} else {
			// TODO(tvoss): Revisit once error propagation strategy is defined. We
			// might want to avoid leaking raw error messages to callers and instead
			// just return a generic error indicating a request ID.
			return nil, err
		}
	}

	if lt := req.GetLatestTime(); lt != nil {
		if ts, err := ptypes.Timestamp(lt); err == nil {
			latest = &ts
		} else {
			// TODO(tvoss): Revisit once error propagation strategy is defined. We
			// might want to avoid leaking raw error messages to callers and instead
			// just return a generic error indicating a request ID.
			return nil, err
		}
	}

	isas, err := s.Store.SearchISAs(ctx, cu, earliest, latest)
	if err != nil {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
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
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, errors.New("missing owner from context")
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

func (s *Server) PatchIdentificationServiceArea(ctx context.Context, req *dspb.PatchIdentificationServiceAreaRequest) (*dspb.PatchIdentificationServiceAreaResponse, error) {
	return nil, nil
}

func (s *Server) PatchSubscription(ctx context.Context, req *dspb.PatchSubscriptionRequest) (*dspb.PatchSubscriptionResponse, error) {
	return nil, nil
}

func (s *Server) PutIdentificationServiceArea(ctx context.Context, req *dspb.PutIdentificationServiceAreaRequest) (*dspb.PutIdentificationServiceAreaResponse, error) {
	return nil, nil
}

func (s *Server) PutSubscription(ctx context.Context, req *dspb.PutSubscriptionRequest) (*dspb.PutSubscriptionResponse, error) {
	return nil, nil
}
