package backend

import (
	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
	"context"
	"database/sql"
)

type DSSStorageInterface interface {
	// Delete reference to an Entity.  USSs should not delete PositionReporting Entities before the end of the last managed flight plus the retention period.
	DeleteUASReporter(context.Context, *dspb.DeleteUASReporterRequest) (*UASReporter, error)
	// Delete a subscription.
	DeleteSubscription(context.Context, *dspb.DeleteSubscriptionRequest) (*Subscription, error)
	// Only PositionReporting Entities shall be visible to clients providing the `dss.read.position_reporting_entities` scope.
	SearchEntityReferences(context.Context, *dspb.SearchUASReportersRequest) ([]*UASReporter, error)
	// Retrieve subscriptions intersecting an area of interest.  Subscription notifications are only triggered by (and contain details of) changes to Entities in the DSS; they do not involve any other data transfer such as remote ID telemetry updates.  Only Subscriptions belonging to the caller are returned.
	SearchSubscriptions(context.Context, *dspb.SearchSubscriptionsRequest) ([]*Subscription, error)
	// Verify the existence/valdity and state of a particular subscription.
	GetSubscription(context.Context, *dspb.GetSubscriptionRequest) (*Subscription, error)
	// Create or update reference to an UASReporter.  Unless otherwise specified, the EntityType of an existing Entity may not be changed.
	PutUASReporter(context.Context, *dspb.PutUASReporterRequest) (*UASReporter, []*Subscription, error)
	// Create or update a subscription.  Subscription notifications are only triggered by (and contain details of) changes to Entities in the DSS; they do not involve any other data transfer such as remote ID telemetry updates.
	//
	// Note that the types of content that should be sent to the created subscription depends on the scope in the provided access token.
	PutSubscription(context.Context, *dspb.PutSubscriptionRequest) (*Subscription, []*UASReporter, error)
}

type CockroachDB struct {
	db *sql.DB
}

func NewCRDB() (*CockroachDB, error) {
	return nil, nil
}

// Delete reference to an Entity.  USSs should not delete PositionReporting Entities before the end of the last managed flight plus the retention period.
func (c *CockroachDB) DeleteUASReporter(context.Context, *dspb.DeleteUASReporterRequest) (*UASReporter, error) {
	return nil, nil
}

// Delete a subscription.
func (c *CockroachDB) DeleteSubscription(context.Context, *dspb.DeleteSubscriptionRequest) (*Subscription, error) {
	return nil, nil
}

// Only PositionReporting Entities shall be visible to clients providing the `dss.read.position_reporting_entities` scope.
func (c *CockroachDB) SearchEntityReferences(context.Context, *dspb.SearchUASReportersRequest) ([]*UASReporter, error) {
	return nil, nil
}

// Retrieve subscriptions intersecting an area of interest.  Subscription notifications are only triggered by (and contain details of) changes to Entities in the DSS; they do not involve any other data transfer such as remote ID telemetry updates.  Only Subscriptions belonging to the caller are returned.
func (c *CockroachDB) SearchSubscriptions(context.Context, *dspb.SearchSubscriptionsRequest) ([]*Subscription, error) {
	return nil, nil
}

func (c *CockroachDB) GetSubscription(context.Context, *dspb.GetSubscriptionRequest) (*Subscription, error) {
	return nil, nil
}

// Create or update reference to a UASReporter.  Unless otherwise specified, the EntityType of an existing Entity may not be changed.
func (c *CockroachDB) PutUASReporter(context.Context, *dspb.PutUASReporterRequest) (*UASReporter, []*Subscription, error) {
	return nil, nil, nil
}

// Create or update a subscription.  Subscription notifications are only triggered by (and contain details of) changes to Entities in the DSS; they do not involve any other data transfer such as remote ID telemetry updates.
//
// Note that the types of content that should be sent to the created subscription depends on the scope in the provided access token.
func (c *CockroachDB) PutSubscription(context.Context, *dspb.PutSubscriptionRequest) (*Subscription, []*UASReporter, error) {
	return nil, nil, nil
}
