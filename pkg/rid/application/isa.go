package application

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
)

// ISAAppInterface provides the interface to the application logic for ISA entities
type ISAAppInterface interface {
	Get(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error)

	// Delete deletes the IdentificationServiceArea identified by "id" and owned by "owner".
	// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
	Delete(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error)

	// InsertISA inserts or updates an ISA.
	Insert(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error)

	// Update
	Update(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error)

	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	Search(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error)
}

// ISAApp is the main implementation of the ISAApp logic.
type ISAApp struct {
	// TODO: don't fully embed the ISA repo once we reduce the complexity in the store.
	// Right now it's "coincidence" that the repo has the same signatures as the App interface
	// but we will want to simplify the repos and add the complexity here.
	repos.ISA
	// TODO:steeling the ISAApp will need access to the Subscription Repo since it touches
	// subs on inserts as well. Probably easiest if it just has the whole set of
	// Repositories
}
