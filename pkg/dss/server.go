package dss

import (
	"context"
	"errors"

	"github.com/steeling/InterUSS-Platform/pkg/dss/auth"
	"github.com/steeling/InterUSS-Platform/pkg/dss/geo"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"

	"github.com/golang/geo/s2"
	"go.uber.org/zap"
)

// Default cell level choices.
//
// The level is chosen such that we operate on cells with an area of ~1km^2.
const (
	DefaultMinimumCellLevel int = 13
	DefaultMaximumCellLevel int = 13
)

var (
	// DefaultRegionCoverer is the default s2.RegionCoverer for mapping areas
	// and extents to s2.CellUnion instances.
	DefaultRegionCoverer = &s2.RegionCoverer{
		MinLevel: DefaultMinimumCellLevel,
		MaxLevel: DefaultMaximumCellLevel,
		// TODO(tvoss): Fine-tune these values.
		LevelMod: 3,
		MaxCells: 10,
	}
)

// Store abstracts interactions with a backend storage layer.
type Store interface {
	// Close closes the store and should release all resources.
	Close() error

	// DeleteIdentificationServiceArea deletes the IdentificationServiceArea identified by "id" and owned by "owner".
	// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
	DeleteIdentificationServiceArea(ctx context.Context, id string, owner string) (*dspb.IdentificationServiceArea, []*dspb.SubscriberToNotify, error)

	// GetSubscription returns the subscription identified by "id".
	GetSubscription(ctx context.Context, id string) (*dspb.Subscription, error)

	// DeleteSubscription deletes the subscription identified by "id" and
	// returns the deleted subscription.
	DeleteSubscription(ctx context.Context, id, version string) (*dspb.Subscription, error)

	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner string) ([]*dspb.Subscription, error)
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

func (ls *loggingStore) DeleteIdentificationServiceArea(ctx context.Context, id string, owner string) (*dspb.IdentificationServiceArea, []*dspb.SubscriberToNotify, error) {
	area, subscriptions, err := ls.next.DeleteIdentificationServiceArea(ctx, id, owner)
	ls.logger.Debug(
		"Store.DeleteIdentificationServiceArea",
		zap.String("id", id),
		zap.String("owner", owner),
		zap.Any("area", area),
		zap.Any("subscriptions", subscriptions),
		zap.Error(err),
	)
	return area, subscriptions, err
}
func (ls *loggingStore) GetSubscription(ctx context.Context, id string) (*dspb.Subscription, error) {
	subscription, err := ls.next.GetSubscription(ctx, id)
	ls.logger.Debug("Store.GetSubscription", zap.String("id", id), zap.Any("subscription", subscription), zap.Error(err))
	return subscription, err
}

func (ls *loggingStore) DeleteSubscription(ctx context.Context, id, version string) (*dspb.Subscription, error) {
	subscription, err := ls.next.DeleteSubscription(ctx, id, version)
	ls.logger.Debug("Store.DeleteSubscription", zap.String("id", id), zap.String("version", version), zap.Any("subscription", subscription), zap.Error(err))
	return subscription, err
}

func (ls *loggingStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner string) ([]*dspb.Subscription, error) {
	subscriptions, err := ls.next.SearchSubscriptions(ctx, cells, owner)
	ls.logger.Debug("Store.SearchSubscriptions", zap.Any("cells", cells), zap.String("owner", owner), zap.Any("subscriptions", subscriptions), zap.Error(err))
	return subscriptions, err
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
	Store   Store
	Coverer *s2.RegionCoverer
	winding geo.WindingOrder
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
	return nil, nil
}

func (s *Server) SearchSubscriptions(ctx context.Context, req *dspb.SearchSubscriptionsRequest) (*dspb.SearchSubscriptionsResponse, error) {
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		// TODO(tvoss): Revisit once error propagation strategy is defined. We
		// might want to avoid leaking raw error messages to callers and instead
		// just return a generic error indicating a request ID.
		return nil, errors.New("missing owner from context")
	}

	cu, err := geo.AreaToCellIDs(req.GetArea(), s.winding, s.Coverer)
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
