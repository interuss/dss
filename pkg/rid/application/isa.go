package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dpjacques/clockwork"
	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
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

	// Insert inserts or updates an ISA.
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
	clock clockwork.Clock
}

// Delete the given ISA
func (a *ISAApp) Delete(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	// We fetch to know whether to return a concurrency error, or a not found error
	old, err := a.ISA.Get(ctx, id)
	switch {
	case err == sql.ErrNoRows || old == nil: // Return a 404 here.
		return nil, nil, dsserr.NotFound(id.String())
	case err != nil:
		return nil, nil, err
	case !version.Empty() && !version.Matches(old.Version):
		return nil, nil, dsserr.VersionMismatch("old version")
	case old.Owner != owner:
		return nil, nil, dsserr.PermissionDenied(fmt.Sprintf("ISA is owned by %s", old.Owner))
	}

	old, subs, err := a.ISA.Delete(ctx, old)
	// TODO: change this to return no error, and a nil object and use that
	// to determine a not found, etc.
	if err == sql.ErrNoRows {
		return nil, nil, dsserr.VersionMismatch("old version")
	}
	return old, subs, err
}

// Insert implments the ISAAppInterface Insert method
func (a *ISAApp) Insert(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	old, err := a.ISA.Get(ctx, isa.ID)
	switch {
	case err == sql.ErrNoRows:
		break
	case err != nil:
		return nil, nil, err
	}

	switch {
	case old == nil && !isa.Version.Empty():
		// The user wants to update an existing ISA, but one wasn't found.
		return nil, nil, dsserr.NotFound(isa.ID.String())
	case old != nil && isa.Version.Empty():
		// The user wants to create a new ISA but it already exists.
		return nil, nil, dsserr.AlreadyExists(isa.ID.String())
	case old != nil && !isa.Version.Matches(old.Version):
		// The user wants to update an ISA but the version doesn't match.
		return nil, nil, dsserr.VersionMismatch("old version")
	case old != nil && old.Owner != isa.Owner:
		return nil, nil, dsserr.PermissionDenied(fmt.Sprintf("ISA is owned by %s", old.Owner))
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := isa.AdjustTimeRange(a.clock.Now(), old); err != nil {
		return nil, nil, err
	}

	return a.ISA.Insert(ctx, isa)
}
