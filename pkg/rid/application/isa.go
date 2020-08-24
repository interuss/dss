package application

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/palantir/stacktrace"
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

func (a *app) GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	repo, err := a.Store.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to interact with store")
	}
	return repo.GetISA(ctx, id)
}

// SearchISAs for ISA within the volume bounds.
func (a *app) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	now := a.clock.Now()
	if earliest == nil || earliest.Before(now) {
		earliest = &now
	}

	repo, err := a.Store.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to interact with store")
	}

	return repo.SearchISAs(ctx, cells, earliest, latest)
}

// DeleteISA the given ISA
func (a *app) DeleteISA(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	var (
		ret  *ridmodels.IdentificationServiceArea
		subs []*ridmodels.Subscription
	)
	// The following will automatically retry TXN retry errors.
	err := a.Store.Transact(ctx, func(repo repos.Repository) error {
		old, err := repo.GetISA(ctx, id)
		switch {
		case err != nil:
			return stacktrace.Propagate(err, "Error getting ISA")
		case old == nil:
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "ISA %s not found", id.String())
		case !version.Matches(old.Version):
			return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
				"ISA currently at version %s but client specified %s", old.Version, version)
		case old.Owner != owner:
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"ISA owned by %s, but %s attempted to delete", old.Owner, owner)
		}

		ret, err = repo.DeleteISA(ctx, old)
		if err != nil {
			return stacktrace.Propagate(err, "Error deleting ISA")
		}

		subs, err = repo.UpdateNotificationIdxsInCells(ctx, old.Cells)
		if err != nil {
			return stacktrace.Propagate(err, "Error updating notification indices")
		}
		return nil
	})
	return ret, subs, err // No need to Propagate this error as this stack layer does not add useful information
}

// InsertISA implments the AppInterface InsertISA method
func (a *app) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	// Validate and perhaps correct StartTime and EndTime.
	if err := isa.AdjustTimeRange(a.clock.Now(), nil); err != nil {
		return nil, nil, stacktrace.Propagate(err, "Error adjusting time range")
	}
	// Update the notification index for both cells removed and added.
	var (
		ret  *ridmodels.IdentificationServiceArea
		subs []*ridmodels.Subscription
	)
	// The following will automatically retry TXN retry errors.
	err := a.Store.Transact(ctx, func(repo repos.Repository) error {
		// ensure it doesn't exist yet
		old, err := repo.GetISA(ctx, isa.ID)
		if err != nil {
			return stacktrace.Propagate(err, "Error getting ISA")
		}
		if old != nil {
			return stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "ISA %s already exists", isa.ID)
		}

		// UpdateNotificationIdxsInCells is done in a Txn along with insert since
		// they are both modifying the db. Insert a susbcription alone does
		// not do this, so that does not need to use a txn (in subscription.go).
		subs, err = repo.UpdateNotificationIdxsInCells(ctx, isa.Cells)
		if err != nil {
			return stacktrace.Propagate(err, "Error updating notification indices")
		}
		ret, err = repo.InsertISA(ctx, isa)
		if err != nil {
			return stacktrace.Propagate(err, "Error inserting ISA")
		}
		return nil
	})
	return ret, subs, err // No need to Propagate this error as this stack layer does not add useful information
}

// UpdateISA implments the AppInterface UpdateISA method
func (a *app) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	// Update the notification index for both cells removed and added.
	var (
		ret  *ridmodels.IdentificationServiceArea
		subs []*ridmodels.Subscription
	)
	// The following will automatically retry TXN retry errors.
	err := a.Store.Transact(ctx, func(repo repos.Repository) error {
		var err error

		old, err := repo.GetISA(ctx, isa.ID)
		switch {
		case err != nil:
			return stacktrace.Propagate(err, "Error getting ISA")
		case old == nil:
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "ISA %s not found", isa.ID)
		case old.Owner != isa.Owner:
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"ISA owned by %s, but %s attempted to modify", old.Owner, isa.Owner)
		case !old.Version.Matches(isa.Version):
			return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
				"ISA currently at version %s but client specified %s", old.Version, isa.Version)
		}
		// Validate and perhaps correct StartTime and EndTime.
		if err := isa.AdjustTimeRange(a.clock.Now(), old); err != nil {
			return stacktrace.Propagate(err, "Error adjusting time range")
		}

		ret, err = repo.UpdateISA(ctx, isa)
		if err != nil {
			return stacktrace.Propagate(err, "Error updating ISA")
		}

		// TODO steeling, we should change this to a Custom type, to obfuscate
		// some of these metrics and prevent us from doing the wrong thing.
		cells := s2.CellUnionFromUnion(old.Cells, isa.Cells)
		geo.Levelify(&cells)
		// UpdateNotificationIdxsInCells is done in a Txn along with insert since
		// they are both modifying the db. Insert a susbcription alone does
		// not do this, so that does not need to use a txn (in subscription.go).
		subs, err = repo.UpdateNotificationIdxsInCells(ctx, cells)
		if err != nil {
			return stacktrace.Propagate(err, "Error updating notification indices")
		}
		return nil
	})

	return ret, subs, err // No need to Propagate this error as this stack layer does not add useful information
}
