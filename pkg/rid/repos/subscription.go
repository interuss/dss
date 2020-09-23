package repos

import (
	"context"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

// Subscription is an interface to a storage layer for the Subscription entity
type Subscription interface {
	// Returns nil, nil if not found
	GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error)

	// DeleteSubscription deletes the IdentificationServiceArea identified by "id" and owned by "owner".
	// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
	// Returns nil, nil if ID, version not found
	DeleteSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// InsertSubscription inserts or updates an ISA.
	InsertSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// UpdateSubscription
	// Returns nil, nil if ID, version not found
	UpdateSubscription(ctx context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// SearchSubscriptions returns all subscriptions ownded by in "cells".
	SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error)

	// SearchSubscriptionsByOwner returns all subscriptions ownded by "owner" in "cells".
	SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error)

	// UpdateNotificationIdxsInCells incremement the notification for each sub in the given cells.
	UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error)

	// MaxSubscriptionCountInCellsByOwner finds, out of a set of cells, the cell with the most subscriptions
	// belonging to the given owner, and returns that number.
	MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error)

	// ListExpiredSubscriptions lists all expired Subscriptions based on writer.
	ListExpiredSubscriptions(ctx context.Context, writer string) ([]*ridmodels.Subscription, error)
}
