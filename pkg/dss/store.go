package dss

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	"github.com/steeling/InterUSS-Platform/pkg/dss/models"
)

type Store interface {
	// Close closes the store and should release all resources.
	Close() error

	GetISA(ctx context.Context, id models.ID) (*models.IdentificationServiceArea, error)

	// Delete deletes the IdentificationServiceArea identified by "id" and owned by "owner".
	// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
	DeleteISA(ctx context.Context, id models.ID, owner models.Owner, version models.Version) (*models.IdentificationServiceArea, []*models.Subscription, error)

	InsertISA(ctx context.Context, isa *models.IdentificationServiceArea) (*models.IdentificationServiceArea, []*models.Subscription, error)

	UpdateISA(ctx context.Context, isa *models.IdentificationServiceArea) (*models.IdentificationServiceArea, []*models.Subscription, error)
	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*models.IdentificationServiceArea, error)

	// GetSubscription returns the subscription identified by "id".
	GetSubscription(ctx context.Context, id models.ID) (*models.Subscription, error)

	// Delete deletes the subscription identified by "id" and
	// returns the deleted subscription.
	DeleteSubscription(ctx context.Context, id models.ID, owner models.Owner, version models.Version) (*models.Subscription, error)

	InsertSubscription(ctx context.Context, s *models.Subscription) (*models.Subscription, error)

	UpdateSubscription(ctx context.Context, s *models.Subscription) (*models.Subscription, error)

	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner models.Owner) ([]*models.Subscription, error)
}

// NewNilStore returns a nil Store instance.
func NewNilStore() Store {
	return nil
}
