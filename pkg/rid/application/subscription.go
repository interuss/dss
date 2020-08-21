package application

import (
	"context"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/palantir/stacktrace"
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

func (a *app) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	repo, err := a.Store.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to interact with store")
	}
	return repo.GetSubscription(ctx, id)
}

func (a *app) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	repo, err := a.Store.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to interact with store")
	}
	return repo.SearchSubscriptionsByOwner(ctx, cells, owner)
}

func (a *app) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	// Validate and perhaps correct StartTime and EndTime.
	if err := s.AdjustTimeRange(a.clock.Now(), nil); err != nil {
		return nil, stacktrace.Propagate(err, "Unable to adjust time range")
	}
	var sub *ridmodels.Subscription
	err := a.Store.Transact(ctx, func(repo repos.Repository) error {

		// ensure it doesn't exist yet
		old, err := repo.GetSubscription(ctx, s.ID)
		if err != nil {
			return stacktrace.Propagate(err, "Error getting Subscription from repo")
		}
		if old != nil {
			return stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "Subscription %s already exists", s.ID)
		}

		// Check the user hasn't created too many subscriptions in this area.
		count, err := repo.MaxSubscriptionCountInCellsByOwner(ctx, s.Cells, s.Owner)
		if err != nil {
			a.logger.Error("Error fetching max subscription count", zap.Error(err))
			return stacktrace.Propagate(err,
				"Failed to fetch subscription count, rejecting request")
		}
		if count >= maxSubscriptionsPerArea {
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.Exhausted, "Too many existing subscriptions in this area already"),
				"%s had %d subscriptions in the area", s.Owner, count)
		}

		sub, err = repo.InsertSubscription(ctx, s)
		if err != nil {
			return stacktrace.Propagate(err, "Error inserting Subscription into repo")
		}

		return nil
	})
	return sub, err
}

// InsertSubscription implements the App InsertSubscription method
func (a *app) UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	var sub *ridmodels.Subscription

	err := a.Store.Transact(ctx, func(repo repos.Repository) error {
		old, err := repo.GetSubscription(ctx, s.ID)
		switch {
		case err != nil:
			return stacktrace.Propagate(err, "Error getting Subscription from repo")
		case old == nil:
			// The user wants to update an existing subscription, but one wasn't found.
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", s.ID.String())
		case !s.Version.Matches(old.Version):
			// The user wants to update a subscription but the version doesn't match.
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", s.Version),
				"Subscription currently at version %s but client specified %s", old.Version, s.Version)
		case old.Owner != s.Owner:
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
				"Subscription owned by %s, but %s attempted to update", old.Owner, s.Owner)
		}
		// Validate and perhaps correct StartTime and EndTime.
		if err := s.AdjustTimeRange(a.clock.Now(), old); err != nil {
			return stacktrace.Propagate(err, "Error adjusting time range")
		}

		// Check the user hasn't created too many subscriptions in this area.
		count, err := repo.MaxSubscriptionCountInCellsByOwner(ctx, s.Cells, s.Owner)
		if err != nil {
			a.logger.Error("Error fetching max subscription count", zap.Error(err))
			return stacktrace.Propagate(err,
				"Failed to fetch subscription count, rejecting request")
		}
		if count >= maxSubscriptionsPerArea {
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.Exhausted, "Too many existing subscriptions in this area already"),
				"%s had %d subscriptions in the area", s.Owner, count)
		}
		sub, err = repo.UpdateSubscription(ctx, s)
		if err != nil {
			return stacktrace.Propagate(err, "Error updating Subscription in repo")
		}
		return nil
	})
	return sub, err
}

// DeleteSubscription deletes the Subscription identified by "id" and owned by "owner".
func (a *app) DeleteSubscription(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, error) {
	var ret *ridmodels.Subscription
	err := a.Store.Transact(ctx, func(repo repos.Repository) error {
		var err error
		old, err := repo.GetSubscription(ctx, id)
		switch {
		case err != nil:
			return stacktrace.Propagate(err, "Error getting Subscription from repo")
		case old == nil:
			return stacktrace.NewErrorWithCode(dsserr.NotFound, "Subscription %s not found", id.String())
		case !version.Matches(old.Version):
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.VersionMismatch, "Subscription version %s is not current", version),
				"Subscription currently at version %s but client specified %s", old.Version, version)
		case old.Owner != owner:
			return stacktrace.Propagate(
				stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Subscription is owned by different client"),
				"Subscription owned by %s, but %s attempted to delete", old.Owner, owner)
		}

		ret, err = repo.DeleteSubscription(ctx, old)
		if err != nil {
			return stacktrace.Propagate(err, "Error deleting Subscription from repo")
		}
		return nil
	})
	return ret, err
}
