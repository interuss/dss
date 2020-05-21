package application

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dpjacques/clockwork"
	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
)

// SubscriptionAppInterface provides the interface to the application logic for Subscription entities
type SubscriptionAppInterface interface {
	Get(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error)

	// Delete deletes the Subscription identified by "id" and owned by "owner".
	// Returns the delete Subscription and all IdentificationServiceAreas affected by the delete.
	Delete(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, error)

	// Insert inserts or updates an Subscription.
	Insert(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// Update
	Update(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// SearchIdentificationServiceAreas returns all IdentificationServiceAreas ownded by "owner" in "cells".
	Search(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error)
}

// SubscriptionApp is the main implementation of the SubscriptionApp logic.
type SubscriptionApp struct {
	// TODO: don't fully embed the Sub repo once we reduce the complexity in the store.
	repos.Subscription
	clock clockwork.Clock
}

// Insert implements the SubscriptionAppInterface Insert method
func (a *SubscriptionApp) Insert(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	old, err := a.Subscription.Get(ctx, s.ID)
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
	if old == nil {
		return a.Subscription.Insert(ctx, s)
	}
	return a.Subscription.Update(ctx, s)
}
