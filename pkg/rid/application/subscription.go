package application

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
)

type IdentificationServiceAreaAppInterface interface {
	Get(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error)

	// Delete deletes the Subscription identified by "id" and owned by "owner".
	// Returns the delete Subscription and all IdentificationServiceAreas affected by the delete.
	Delete(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, []*ridmodels.IdentificationServiceArea, error)

	// InsertISA inserts or updates an ISA.
	Insert(ctx context.Context, isa *ridmodels.Subscription) (*ridmodels.Subscription, []*ridmodels.IdentificationServiceArea, error)

	// Update
	Update(ctx context.Context, isa *ridmodels.Subscription) (*ridmodels.Subscription, []*ridmodels.IdentificationServiceArea, error)

	// SearchIdentificationServiceAreas returns all IdentificationServiceAreas ownded by "owner" in "cells".
	Search(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.Subscription, error)
}

type SubscriptionApp struct {
	ir repos.ISA
	sr repos.Subscription
}
