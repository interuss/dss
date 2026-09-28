package application

import (
	"context"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
)

// SubscriptionApp provides the interface to the application logic for Subscription entities
// AppInterface provides the interface to the application logic for ISA entities
// Note that there is no need for the applciation layer to have the same API as
// the repo layer.
type SubscriptionApp interface {
	GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error)

	// SearchSubscriptionsByOwner returns all IdentificationServiceAreas ownded by "owner" in "cells".
	SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error)
}

func (a *app) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	repo, err := a.store.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to interact with store")
	}
	return repo.GetSubscription(ctx, id)
}

func (a *app) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	repo, err := a.store.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to interact with store")
	}
	return repo.SearchSubscriptionsByOwner(ctx, cells, owner)
}
