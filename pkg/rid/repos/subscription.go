package repos

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

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
	Search(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.Subscription, error)
}
