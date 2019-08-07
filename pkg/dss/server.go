package dss

import (
	"context"
	"errors"
	"time"

	"github.com/golang/geo/s2"
	"github.com/golang/protobuf/ptypes"
	"github.com/steeling/InterUSS-Platform/pkg/dss/auth"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
)

var (
	WriteISAScope = "dss.write.identification_service_areas"
	ReadISAScope  = "dss.read.identification_service_areas"
)

type Store interface {
	// Close closes the store and should release all resources.
	Close() error

	// DeleteIdentificationServiceArea deletes the IdentificationServiceArea
	// identified by "id" and owned by "owner". Returns the delete
	// IdentificationServiceArea and all Subscriptions affected by the delete.
	DeleteIdentificationServiceArea(ctx context.Context, id string, owner string) (*dspb.IdentificationServiceArea, []*dspb.SubscriberToNotify, error)

	// SearchIdentificationServiceAreas searches IdentificationServiceArea
	// instances that intersect with "cells" and, if set, the temporal volume
	// defined by "earliest" and "latest".
	SearchIdentificationServiceAreas(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*dspb.IdentificationServiceArea, error)

	// GetSubscription returns the subscription identified by "id".
	GetSubscription(ctx context.Context, id string) (*dspb.Subscription, error)

	// DeleteSubscription deletes the subscription identified by "id" and
	// returns the deleted subscription.
	DeleteSubscription(ctx context.Context, id, version string) (*dspb.Subscription, error)

	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner string) ([]*dspb.Subscription, error)
}

// NewNilStore returns a nil Store instance.
func NewNilStore() Store {
	return nil
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

func (s *Server) DeleteIdentificationServiceArea(ctx context.Context, req *dspb.DeleteIdentificationServiceAreaRequest) (*dspb.DeleteIdentificationServiceAreaResponse, error) {
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, errors.New("missing owner from context")
	}

	isa, subscribers, err := s.Store.DeleteIdentificationServiceArea(ctx, req.GetId(), owner)
	if err != nil {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, err
	}

	return &dspb.DeleteIdentificationServiceAreaResponse{
		ServiceArea: isa,
		Subscribers: subscribers,
	}, nil
}

func (s *Server) DeleteSubscription(ctx context.Context, req *dspb.DeleteSubscriptionRequest) (*dspb.DeleteSubscriptionResponse, error) {
	subscription, err := s.Store.DeleteSubscription(ctx, req.GetId(), req.GetVersion())
	if err != nil {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, err
	}
	return &dspb.DeleteSubscriptionResponse{
		Subscription: subscription,
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

	serviceAreas, err := s.Store.SearchIdentificationServiceAreas(ctx, cu, earliest, latest)
	if err != nil {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, err
	}

	return &dspb.SearchIdentificationServiceAreasResponse{
		ServiceAreas: serviceAreas,
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

	return &dspb.SearchSubscriptionsResponse{
		Subscriptions: subscriptions,
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
	return &dspb.GetSubscriptionResponse{
		Subscription: subscription,
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
