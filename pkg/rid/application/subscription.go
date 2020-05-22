package application

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

// SubscriptionApp provides the interface to the application logic for Subscription entities
// AppInterface provides the interface to the application logic for ISA entities
// Note that there is no need for the applciation layer to have the same API as
// the repo layer.
type SubscriptionApp interface {
	GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error)

	// DeleteSubscription deletes the Subscription identified by "id" and owned by "owner".
	// Returns the delete Subscription and all IdentificationServiceAreas affected by the delete.
	DeleteSubscription(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, error)

	// InsertSubscription inserts or updates an Subscription.
	InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// UpdateSubscription
	UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// SearchSubscriptionsByOwner returns all IdentificationServiceAreas ownded by "owner" in "cells".
	SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error)
}

// InsertSubscription implements the App InsertSubscription method
func (a *app) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	old, err := a.Subscription.GetSubscription(ctx, s.ID)
	switch {
	case err == sql.ErrNoRows:
		break
	case err != nil:
		return nil, err
	}
	switch {
	case old == nil && !s.Version.Empty():
		// The user wants to update an existing subscription, but one wasn't found.
		return nil, dsserr.NotFound(s.ID.String())
	case old != nil && s.Version.Empty():
		// The user wants to create a new subscription but it already exists.
		return nil, dsserr.AlreadyExists(s.ID.String())
	case old != nil && !s.Version.Matches(old.Version):
		// The user wants to update a subscription but the version doesn't match.
		return nil, dsserr.VersionMismatch("old version")
	case old != nil && old.Owner != s.Owner:
		return nil, dsserr.PermissionDenied(fmt.Sprintf("s is owned by %s", old.Owner))
	}
	// Validate and perhaps correct StartTime and EndTime.
	if err := s.AdjustTimeRange(a.clock.Now(), old); err != nil {
		return nil, err
	}

	return a.Subscription.InsertSubscription(ctx, s)
}

// DeleteSubscription deletes the Subscription identified by "id" and owned by "owner".
func (a *app) DeleteSubscription(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, error) {

	old, err := a.Subscription.GetSubscription(ctx, id)
	switch {
	case err == sql.ErrNoRows || old == nil:
		return nil, dsserr.NotFound(id.String())
	case err != nil:
		return nil, err
	case old.Owner != owner:
		return nil, dsserr.PermissionDenied(fmt.Sprintf("ISA is owned by %s", old.Owner))
	}
	old, err = a.Subscription.DeleteSubscription(ctx, old)
	if err == sql.ErrNoRows {
		return nil, dsserr.VersionMismatch("old version")
	}
	return old, err
}
