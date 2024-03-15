package repos

import (
	"context"
	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

// Subscriptions enables operations on a list of Subscriptions.
type Subscriptions []*scdmodels.Subscription

// OperationalIntent abstracts operational intent-specific interactions with the backing repository.
type OperationalIntent interface {
	// GetOperationalIntent returns the operation identified by "id".
	GetOperationalIntent(ctx context.Context, id dssmodels.ID) (*scdmodels.OperationalIntent, error)

	// DeleteOperationalIntent deletes the operation identified by "id".
	DeleteOperationalIntent(ctx context.Context, id dssmodels.ID) error

	// UpsertOperationalIntent inserts or updates an operation into the store.
	UpsertOperationalIntent(ctx context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error)

	// SearchOperationalIntents returns all operations intersecting "v4d".
	SearchOperationalIntents(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.OperationalIntent, error)

	// GetDependentOperationalIntents returns IDs of all operations dependent on
	// subscription identified by "subscriptionID".
	GetDependentOperationalIntents(ctx context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error)
}

// Subscription abstracts subscription-specific interactions with the backing repository.
type Subscription interface {
	// SearchSubscriptions returns all Subscriptions in "v4d".
	SearchSubscriptions(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Subscription, error)

	// GetSubscription returns the Subscription referenced by id, or nil and no
	// error if the Subscription doesn't exist
	GetSubscription(ctx context.Context, id dssmodels.ID) (*scdmodels.Subscription, error)

	// UpsertSubscription upserts sub into the store and returns the result
	// subscription.
	UpsertSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error)

	// DeleteSubscription deletes a Subscription from the store and returns the
	// deleted subscription.  Returns an error if the Subscription does not
	// exist.
	DeleteSubscription(ctx context.Context, id dssmodels.ID) error

	// IncrementNotificationIndices increments the notification index of each
	// specified Subscription and returns the resulting corresponding
	// notification indices.
	IncrementNotificationIndices(ctx context.Context, subscriptionIds []dssmodels.ID) ([]int, error)

	// LockSubscriptionsOnCells locks the subscriptions of interest on specific cells.
	LockSubscriptionsOnCells(ctx context.Context, cells s2.CellUnion) error
}

type UssAvailability interface {
	GetUssAvailability(ctx context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error)

	UpsertUssAvailability(ctx context.Context, ussa *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error)
}

// repos.Constraint abstracts constraint-specific interactions with the backing store.
type Constraint interface {
	// SearchConstraints returns all Constraints in "v4d".
	SearchConstraints(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Constraint, error)

	// GetConstraint returns the Constraint referenced by id, or
	// (nil, sql.ErrNoRows) if the Constraint doesn't exist
	GetConstraint(ctx context.Context, id dssmodels.ID) (*scdmodels.Constraint, error)

	// UpsertConstraint upserts "constraint" into the store.
	UpsertConstraint(ctx context.Context, constraint *scdmodels.Constraint) (*scdmodels.Constraint, error)

	// DeleteConstraint deletes a Constraint from the store and returns the
	// deleted subscription.  Returns nil and an error if the Constraint does
	// not exist.
	DeleteConstraint(ctx context.Context, id dssmodels.ID) error
}

// DssReport takes care of handling a DSS report.
type DssReport interface {
	// HandleDssReport handles a DSS report request. Returns the error report passed in 'req' after having set its identifier.
	HandleDssReport(ctx context.Context, req *restapi.MakeDssReportRequest) (*restapi.ErrorReport, error)
}

// Repository aggregates all SCD-specific repo interfaces.
// Note that 'DssReport' is not yet part of it, while we figure exactly how we want to handle DSS reports
type Repository interface {
	OperationalIntent
	Subscription
	Constraint
	UssAvailability
}

// IncrementNotificationIndices is a utility function that extracts the IDs from
// a list of Subscriptions before calling the underlying repo function, and then
// updates the Subscription objects with the new notification indices.
func (subs Subscriptions) IncrementNotificationIndices(ctx context.Context, r Repository) error {
	subIds := make([]dssmodels.ID, len(subs))
	for i, sub := range subs {
		subIds[i] = sub.ID
	}
	newIndices, err := r.IncrementNotificationIndices(ctx, subIds)
	if err != nil {
		return err
	}
	for i, newIndex := range newIndices {
		subs[i].NotificationIndex = newIndex
	}
	return nil
}
