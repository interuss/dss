package application

import (
	"context"
	"fmt"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	// Defined in requirement DSS0030.
	maxSubscriptionsPerArea = 10
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

func (a *app) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	// Validate and perhaps correct StartTime and EndTime.
	if err := s.AdjustTimeRange(a.clock.Now(), nil); err != nil {
		return nil, err
	}

	// Check the user hasn't created too many subscriptions in this area.
	count, err := a.Transactor.MaxSubscriptionCountInCellsByOwner(ctx, s.Cells, s.Owner)
	if err != nil {
		a.logger.Error("Error fetching max subscription count", zap.Error(err))
		return nil, dsserr.Internal(
			"failed to fetch subscription count, rejecting request")
	}
	if count >= maxSubscriptionsPerArea {
		return nil, dsserr.Exhausted(
			"too many existing subscriptions in this area already")
	}

	sub, err := a.Transactor.InsertSubscription(ctx, s)
	if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
		return dsserr.AlreadyExists("Sub with ID: %s already exists", isa.ID)
	}
	return sub, err
}

// InsertSubscription implements the App InsertSubscription method
func (a *app) UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	old, err := a.Transactor.GetSubscription(ctx, s.ID)
	switch {
	case err != nil:
		return nil, err
	case old == nil:
		// The user wants to update an existing subscription, but one wasn't found.
		return nil, dsserr.NotFound(s.ID.String())
	case !s.Version.Matches(old.Version):
		// The user wants to update a subscription but the version doesn't match.
		return nil, dsserr.VersionMismatch("old version")
	case old.Owner != s.Owner:
		return nil, dsserr.PermissionDenied(fmt.Sprintf("s is owned by %s", old.Owner))
	}
	// Validate and perhaps correct StartTime and EndTime.
	if err := s.AdjustTimeRange(a.clock.Now(), old); err != nil {
		return nil, err
	}

	// Check the user hasn't created too many subscriptions in this area.
	count, err := a.Transactor.MaxSubscriptionCountInCellsByOwner(ctx, s.Cells, s.Owner)
	if err != nil {
		a.logger.Error("Error fetching max subscription count", zap.Error(err))
		return nil, dsserr.Internal(
			"failed to fetch subscription count, rejecting request")
	}
	if count >= maxSubscriptionsPerArea {
		return nil, dsserr.Exhausted(
			"too many existing subscriptions in this area already")
	}

	return a.Transactor.UpdateSubscription(ctx, s)
}

// DeleteSubscription deletes the Subscription identified by "id" and owned by "owner".
func (a *app) DeleteSubscription(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, error) {
	var old *ridmodels.Subscription
	err := a.Transactor.InTxnRetrier(ctx, func(repo repos.Repository) error {
		var err error
		old, err = repo.UnsafeDeleteSubscription(ctx, old)
		switch {
		case err != nil:
			return err
		case old == nil: // Return a 404 here.
			return dsserr.NotFound(id.String(), version.String())
		case !version.Matches(old.Version):
			return dsserr.VersionMismatch("old version")
		case old.Owner != owner:
			return dsserr.PermissionDenied(fmt.Sprintf("Sub is owned by %s", old.Owner))
		}
		return nil
	})
	return old, err
}
