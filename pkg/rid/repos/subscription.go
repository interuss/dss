package repos

import (
	"context"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

// Subscription is an interface to a storage layer for the Subscription entity
type Subscription interface {
	Get(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error)

	// Delete deletes the IdentificationServiceArea identified by "id" and owned by "owner".
	// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
	Delete(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, error)

	// InsertISA inserts or updates an ISA.
	Insert(ctx context.Context, isa *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// Update
	Update(ctx context.Context, isa *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	Search(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error)
}
