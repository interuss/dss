package dss

import (
	"context"

	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
	"go.uber.org/zap"
)

// Store abstracts interactions with a backend storage layer.
type Store interface {
	// Close closes the store and should release all resources.
	Close() error

	// GetSubscription returns the subscription identified by "id".
	GetSubscription(ctx context.Context, id string) (*dspb.Subscription, error)

	// DeleteSubscription deletes the subscription identified by "id" and
	// returns the deleted subscription.
	DeleteSubscription(ctx context.Context, id string) (*dspb.Subscription, error)
}

type loggingStore struct {
	logger *zap.Logger
	next   Store
}

func (ls *loggingStore) Close() error {
	err := ls.next.Close()
	ls.logger.Debug("Store.Close", zap.Error(err))
	return err
}

func (ls *loggingStore) GetSubscription(ctx context.Context, id string) (*dspb.Subscription, error) {
	subscription, err := ls.next.GetSubscription(ctx, id)
	ls.logger.Debug("Store.GetSubscription", zap.String("id", id), zap.Any("subscription", subscription), zap.Error(err))
	return subscription, err
}

func (ls *loggingStore) DeleteSubscription(ctx context.Context, id string) (*dspb.Subscription, error) {
	subscription, err := ls.next.DeleteSubscription(ctx, id)
	ls.logger.Debug("Store.DeleteSubscription", zap.String("id", id), zap.Any("subscription", subscription), zap.Error(err))
	return subscription, err
}

// DecorateLogging decorates store with logging at debug level.
func DecorateLogging(logger *zap.Logger, store Store) Store {
	return &loggingStore{
		logger: logger,
		next:   store,
	}
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

func (s *Server) PutIdentificationServiceArea(ctx context.Context, req *dspb.PutIdentificationServiceAreaRequest) (*dspb.PutIdentificationServiceAreaResponse, error) {
	return nil, nil
}

func (s *Server) PutSubscription(ctx context.Context, req *dspb.PutSubscriptionRequest) (*dspb.PutSubscriptionResponse, error) {
	return nil, nil
}
