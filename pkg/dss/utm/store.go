package utm

import (
	"context"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/dss/utm/models"
)

// Store abstracts interactions with a backing data store.
type Store interface {
	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner models.Owner) ([]*models.Subscription, error)

	// GetSubscription returns the subscription referenced by id.
	GetSubscription(ctx context.Context, id models.ID, owner models.Owner) (*models.Subscription, error)

	// InsertSubscription inserts sub into the store and returns the result
	// subscription.
	InsertSubscription(ctx context.Context, sub *models.Subscription, owner models.Owner) (*models.Subscription, error)

	// DeleteSubscription deletes a sub from the store and returns the deleted subscription.
	DeleteSubscription(ctx context.Context, id models.ID, owner models.Owner) (*models.Subscription, error)
}
