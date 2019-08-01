package dss

import (
	"context"

	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
)

// Store abstracts interactions with a backend storage layer.
type Store interface {
	// Close closes the store and should release all resources.
	Close() error

	// DeleteSubscription deletes the subscription identified by "id" and
	// returns the deleted subscription.
	DeleteSubscription(ctx context.Context, id string) (*dspb.Subscription, error)
}

// NewNilStore returns a nil Store instance.
func NewNilStore() Store {
	return nil
}

// Server implements dssproto.DiscoveryAndSynchronizationService.
type Server struct {
	Store Store
}

func (s *Server) DeleteIdentificationServiceArea(ctx context.Context, req *dspb.DeleteIdentificationServiceAreaRequest) (*dspb.DeleteIdentificationServiceAreaResponse, error) {
	return nil, nil
}

func (s *Server) DeleteSubscription(ctx context.Context, req *dspb.DeleteSubscriptionRequest) (*dspb.DeleteSubscriptionResponse, error) {
	subscription, err := s.Store.DeleteSubscription(ctx, req.GetId())
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
	return nil, nil
}

func (s *Server) SearchSubscriptions(ctx context.Context, req *dspb.SearchSubscriptionsRequest) (*dspb.SearchSubscriptionsResponse, error) {
	return nil, nil
}

func (s *Server) GetSubscription(ctx context.Context, req *dspb.GetSubscriptionRequest) (*dspb.GetSubscriptionResponse, error) {
	return nil, nil
}

func (s *Server) PutIdentificationServiceArea(ctx context.Context, req *dspb.PutIdentificationServiceAreaRequest) (*dspb.PutIdentificationServiceAreaResponse, error) {
	return nil, nil
}

func (s *Server) PutSubscription(ctx context.Context, req *dspb.PutSubscriptionRequest) (*dspb.PutSubscriptionResponse, error) {
	return nil, nil
}
