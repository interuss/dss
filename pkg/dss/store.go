package dss

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
	"go.uber.org/zap"
)

type Store interface {
	// Close closes the store and should release all resources.
	Close() error

	// SearchIdentificationServiceAreas searches IdentificationServiceArea
	// instances that intersect with "cells" and, if set, the temporal volume
	// defined by "earliest" and "latest".
	SearchIdentificationServiceAreas(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*dspb.IdentificationServiceArea, error)

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

// NewNilStore returns a nil Store instance.
func NewNilStore() Store {
	return nil
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

func (ls *loggingStore) SearchIdentificationServiceAreas(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*dspb.IdentificationServiceArea, error) {
	areas, err := ls.next.SearchIdentificationServiceAreas(ctx, cells, earliest, latest)
	ls.logger.Debug(
		"Store.SearchIdentificationServiceAreas",
		zap.Any("cells", cells),
		zap.Any("earliest", earliest),
		zap.Any("latest", latest),
		zap.Any("areas", areas),
		zap.Error(err),
	)
	return areas, err
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

// DecorateWithLogging decorates store with logging at debug level.
func DecorateWithLogging(logger *zap.Logger, store Store) Store {
	return &loggingStore{
		logger: logger,
		next:   store,
	}
}
