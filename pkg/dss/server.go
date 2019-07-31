package dss

import (
	"context"

	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
)

// Store abstracts interactions with a backend storage layer.
type Store interface {
	// Close closes the store and should release all resources.
	Close() error
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
	return nil, nil
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
