package dss

import (
	"context"
	"errors"
	"time"

	"github.com/golang/geo/s2"
	"github.com/steeling/InterUSS-Platform/pkg/dss/models"
	"go.uber.org/zap"
)

var (
	ErrAlreadyExists   = errors.New("resource already exists")
	ErrVersionMismatch = errors.New("version mismatch for resource")
	ErrNotFound        = errors.New("resource not found")
	ErrBadRequest      = errors.New("bad request")
)

type Store interface {
	// Close closes the store and should release all resources.
	Close() error

	// Delete deletes the IdentificationServiceArea identified by "id" and owned by "owner".
	// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
	DeleteISA(ctx context.Context, id, owner, version string) (*models.IdentificationServiceArea, []*models.Subscription, error)

	InsertISA(ctx context.Context, isa *models.IdentificationServiceArea) (*models.IdentificationServiceArea, []*models.Subscription, error)

	UpdateISA(ctx context.Context, isa *models.IdentificationServiceArea) (*models.IdentificationServiceArea, []*models.Subscription, error)
	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*models.IdentificationServiceArea, error)

	// GetSubscription returns the subscription identified by "id".
	GetSubscription(ctx context.Context, id string) (*models.Subscription, error)

	// Delete deletes the subscription identified by "id" and
	// returns the deleted subscription.
	DeleteSubscription(ctx context.Context, id, owner, version string) (*models.Subscription, error)

	InsertSubscription(ctx context.Context, s *models.Subscription) (*models.Subscription, error)

	UpdateSubscription(ctx context.Context, s *models.Subscription) (*models.Subscription, error)

	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner string) ([]*models.Subscription, error)
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

func (ls *loggingStore) DeleteISA(ctx context.Context, id, owner, version string) (*models.IdentificationServiceArea, []*models.Subscription, error) {
	area, subscriptions, err := ls.next.DeleteISA(ctx, id, owner, version)
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

func (ls *loggingStore) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*models.IdentificationServiceArea, error) {
	areas, err := ls.next.SearchISAs(ctx, cells, earliest, latest)
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

func (ls *loggingStore) GetSubscription(ctx context.Context, id string) (*models.Subscription, error) {
	subscription, err := ls.next.GetSubscription(ctx, id)
	ls.logger.Debug("Store.GetSubscription", zap.String("id", id), zap.Any("subscription", subscription), zap.Error(err))
	return subscription, err
}

func (ls *loggingStore) DeleteSubscription(ctx context.Context, id, owner, version string) (*models.Subscription, error) {
	subscription, err := ls.next.DeleteSubscription(ctx, id, owner, version)
	ls.logger.Debug("Store.DeleteSubscription", zap.String("id", id), zap.String("version", version), zap.Any("subscription", subscription), zap.Error(err))
	return subscription, err
}

func (ls *loggingStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner string) ([]*models.Subscription, error) {
	subscriptions, err := ls.next.SearchSubscriptions(ctx, cells, owner)
	ls.logger.Debug("Store.SearchSubscriptions", zap.Any("cells", cells), zap.String("owner", owner), zap.Any("subscriptions", subscriptions), zap.Error(err))
	return subscriptions, err
}
