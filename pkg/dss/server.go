package dss

import (
	"context"

	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
)

// StorageInterface abstracts interactions with a backend storage layer.
type StorageInterface interface{}

// NewNilStorageInterface returns a nil StorageInterface.
func NewNilStorageInterface() StorageInterface {
	return nil
}

// Server implements dssproto.DiscoveryAndSynchronizationService.
type Server struct {
	store StorageInterface
}

// NewServer returns a new Server instance using store or an error in case of
// issues.
func NewServer(store StorageInterface) (*Server, error) {
	return &Server{
		store: store,
	}, nil
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
