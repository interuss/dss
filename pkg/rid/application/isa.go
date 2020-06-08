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
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/lib/pq"
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

	return a.Transactor.SearchISAs(ctx, cells, earliest, latest)
}

// DeleteISA the given ISA
func (a *app) DeleteISA(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	var (
		old  *ridmodels.IdentificationServiceArea
		subs []*ridmodels.Subscription
	)
	// The following will automatically retry TXN retry errors.
	err := a.Transactor.InTxnRetrier(ctx, func(repo repos.Repository) error {
		var err error

		old, err = repo.UnsafeDeleteISA(ctx, old)
		switch {
		case err == sql.ErrNoRows: // Return a 404 here.
			return dsserr.NotFound(id.String(), version.String())
		case err != nil:
			return err
		case !version.Matches(old.Version):
			return dsserr.VersionMismatch("old version")
		case old.Owner != owner:
			return dsserr.PermissionDenied(fmt.Sprintf("ISA is owned by %s", old.Owner))
		}

		subs, err = repo.UpdateNotificationIdxsInCells(ctx, old.Cells)
		return err
	})
	return old, subs, err
}

// InsertISA implments the AppInterface InsertISA method
func (a *app) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	// Validate and perhaps correct StartTime and EndTime.
	if err := isa.AdjustTimeRange(a.clock.Now(), nil); err != nil {
		return nil, nil, err
	}
	// Update the notification index for both cells removed and added.
	var (
		ret  *ridmodels.IdentificationServiceArea
		subs []*ridmodels.Subscription
	)
	// The following will automatically retry TXN retry errors.
	err = a.Transactor.InTxnRetrier(ctx, func(repo repos.Repository) error {
		var err error
		// UpdateNotificationIdxsInCells is done in a Txn along with insert since
		// they are both modifying the db. Insert a susbcription alone does
		// not do this, so that does not need to use a txn (in subscription.go).
		subs, err = a.Transactor.UpdateNotificationIdxsInCells(ctx, isa.Cells)
		if err != nil {
			return err
		}
		ret, err = a.Transactor.InsertISA(ctx, isa)
		if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
			return dsserr.AlreadyExists("ISA with ID: %s already exists", isa.ID)
		}
		if err != nil {
			return err
		}
		return nil
	})
	return ret, subs, err
}

// UpdateISA implments the AppInterface InsertISA method
func (a *app) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	old, err := a.Transactor.GetISA(ctx, isa.ID)
	if err != nil {
		return nil, nil, err
	}

	if old.Owner != isa.Owner {
		return nil, nil, dsserr.PermissionDenied(fmt.Sprintf("ISA is owned by %s", old.Owner))
	}

	// Validate and perhaps correct StartTime and EndTime.
	// TODO: recommended to explicitly force the user to pass in the correct time,
	// instead of changing it on them.
	if err := isa.AdjustTimeRange(a.clock.Now(), old); err != nil {
		return nil, nil, err
	}
	// Update the notification index for both cells removed and added.
	var (
		ret  *ridmodels.IdentificationServiceArea
		subs []*ridmodels.Subscription
	)
	// The following will automatically retry TXN retry errors.
	err = a.Transactor.InTxnRetrier(ctx, func(repo repos.Repository) error {
		var err error
		// TODO steeling, we should change this to a Custom type, to obfuscate
		// some of these metrics and prevent us from doing the wrong thing.
		cells := s2.CellUnionFromUnion(old.Cells, isa.Cells)
		geo.Levelify(&cells)
		// UpdateNotificationIdxsInCells is done in a Txn along with insert since
		// they are both modifying the db. Insert a susbcription alone does
		// not do this, so that does not need to use a txn (in subscription.go).
		subs, err = a.Transactor.UpdateNotificationIdxsInCells(ctx, cells)
		if err != nil {
			return err
		}
		ret, err = a.Transactor.InsertISA(ctx, isa)
		return err
	})
	// This can happen if either the version was changed, or the entity was
	// deleted. The grpc error code for already exists maps to an HTTP 409
	// CONFLICT code.
	if ret == nil {
		return nil, nil, dsserr.AlreadyExists("this ISA was either updated or deleted")
	}
	return ret, subs, err
}
