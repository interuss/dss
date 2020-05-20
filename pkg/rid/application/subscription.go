package application

import (
	"context"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
)

type SubscriptionAppInterface interface {
	Get(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error)

	// Delete deletes the Subscription identified by "id" and owned by "owner".
	// Returns the delete Subscription and all IdentificationServiceAreas affected by the delete.
	Delete(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, error)

	// InsertISA inserts or updates an ISA.
	Insert(ctx context.Context, isa *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// Update
	Update(ctx context.Context, isa *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// SearchIdentificationServiceAreas returns all IdentificationServiceAreas ownded by "owner" in "cells".
	Search(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error)
}

type SubscriptionApp struct {
	// TODO: don't fully embed the Sub repo once we reduce the complexity in the store.
	repos.Subscription
}
