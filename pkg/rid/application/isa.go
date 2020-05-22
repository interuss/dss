package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

// AppInterface provides the interface to the application logic for ISA entities
// Note that there is no need for the applciation layer to have the same API as
// the repo layer.
type ISAApp interface {
	GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error)

	// DeleteISA deletes the IdentificationServiceArea identified by "id" and owned by "owner".
	// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
	DeleteISA(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error)

	// InsertISA inserts or updates an ISA.
	InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error)

	// UpdateISA
	UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error)

	// SearchISAs returns all subscriptions ownded by "owner" in "cells".
	SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error)
}

// SearchISAs for ISA within the volume bounds.
func (a *app) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	now := a.clock.Now()
	if earliest == nil || earliest.Before(now) {
		earliest = &now
	}

	return a.ISA.SearchISAs(ctx, cells, earliest, latest)
}

// DeleteISA the given ISA
func (a *app) DeleteISA(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	// We fetch to know whether to return a concurrency error, or a not found error
	old, err := a.ISA.GetISA(ctx, id)
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

	// TODO(steeling) do this in a txn.
	subs, err := a.Subscription.UpdateNotificationIdxsInCells(ctx, old.Cells)
	if err != nil {
		return nil, nil, err
	}

	isa, err := a.ISA.DeleteISA(ctx, old)
	// TODO: change this to return no error, and a nil object and use that
	// to determine a not found, etc.
	if err == sql.ErrNoRows {
		return nil, nil, dsserr.VersionMismatch("old version")
	}

	return isa, subs, err
}

// InsertISA implments the AppInterface InsertISA method
func (a *app) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	old, err := a.ISA.GetISA(ctx, isa.ID)
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
	// TODO(steeling) do this in a txn.
	// Update the notification index for both cells removed and added.
	cells := isa.Cells
	if old != nil {
		// TODO steeling, we should change this to a Custom type, to obfuscate
		// some of these metrics and prevent us from doing the wrong thing.
		cells = s2.CellUnionFromUnion(old.Cells, isa.Cells)
		geo.Levelify(&cells)
	}
	subs, err := a.Subscription.UpdateNotificationIdxsInCells(ctx, cells)
	if err != nil {
		return nil, nil, err
	}
	isa, err = a.ISA.InsertISA(ctx, isa)
	if err != nil {
		return nil, nil, err
	}
	return isa, subs, nil
}

// UpdateISA implments the AppInterface InsertISA method
func (a *app) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	return nil, nil, dsserr.Internal("not yet implemented")
}
