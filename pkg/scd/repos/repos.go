package repos

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

// Operation abstracts operation-specific interactions with the backing repository.
type Operation interface {
	// GetOperation returns the operation identified by "id".
	GetOperation(ctx context.Context, id scdmodels.ID) (*scdmodels.Operation, error)

	// DeleteOperation deletes the operation identified by "id" and owned by "owner".
	// Returns the deleted Operation and all Subscriptions affected by the delete.
	DeleteOperation(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Operation, []*scdmodels.Subscription, error)

	// UpsertOperation inserts or updates an operation using key as a fencing
	// token. If operation does not reference an existing subscription, an
	// implicit subscription with parameters notifySubscriptionForConstraints
	// and subscriptionBaseURL is created.
	UpsertOperation(ctx context.Context, operation *scdmodels.Operation, key []scdmodels.OVN) (*scdmodels.Operation, []*scdmodels.Subscription, error)

	// SearchOperations returns all operations ownded by "owner" intersecting "v4d".
	SearchOperations(ctx context.Context, v4d *dssmodels.Volume4D, owner dssmodels.Owner) ([]*scdmodels.Operation, error)
}

// Subscription abstracts subscription-specific interactions with the backing repository.
type Subscription interface {
	// SearchSubscriptions returns all Subscriptions in "v4d".
	SearchSubscriptions(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error)

	// GetSubscription returns the Subscription referenced by id, or nil if the
	// Subscription doesn't exist
	GetSubscription(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Subscription, error)

	// UpsertSubscription upserts sub into the store and returns the result
	// subscription.
	UpsertSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, []*scdmodels.Operation, error)

	// DeleteSubscription deletes a Subscription from the store and returns the
	// deleted subscription.  Returns nil and an error if the Subscription does
	// not exist, or is owned by someone other than the specified owner.
	DeleteSubscription(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner, version scdmodels.Version) (*scdmodels.Subscription, error)
}

// Repository aggregates all entity-specific repo interfaces.
type Repository interface {
	Operation
	Subscription
}
