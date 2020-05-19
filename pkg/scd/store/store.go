package store

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

// OperationStore abstracts operation-specific interactions with the backing data store.
type OperationStore interface {
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
	SearchOperations(ctx context.Context, v4d *scdmodels.Volume4D, owner dssmodels.Owner) ([]*scdmodels.Operation, error)
}

// SubscriptionStore abstracts subscription-specific interactions with the backing data store.
type SubscriptionStore interface {
	// SearchSubscriptions returns all Subscriptions owned by "owner" in "cells".
	SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*scdmodels.Subscription, error)

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

// Store abstracts interactions with a backing data store.
type Store interface {
	OperationStore
	SubscriptionStore
}

var (
	_ Store = &DummyStore{}
)

// DummyStore implements Store interface entirely with error-free no-ops
type DummyStore struct {
}

func makeDummyOperation(id scdmodels.ID) *scdmodels.Operation {
	var (
		altLo = float32(11235)
		altHi = float32(81321)
		start = time.Now()
		end   = start.Add(2 * time.Minute)
	)

	return &scdmodels.Operation{
		ID:            id,
		Version:       314,
		OVN:           scdmodels.NewOVNFromTime(time.Now(), id.String()),
		USSBaseURL:    "https://exampleuss.com/utm",
		AltitudeLower: &altLo,
		AltitudeUpper: &altHi,
		StartTime:     &start,
		EndTime:       &end,
		State:         scdmodels.OperationStateAccepted,
	}

}

// makeDummySubscription returns a dummy subscription instance with ID id.
func makeDummySubscription(id scdmodels.ID) *scdmodels.Subscription {
	altLo := float32(11235)
	altHi := float32(81321)
	result := &scdmodels.Subscription{
		ID:                   id,
		Version:              314,
		NotificationIndex:    123,
		BaseURL:              "https://exampleuss.com/utm",
		AltitudeLo:           &altLo,
		AltitudeHi:           &altHi,
		NotifyForOperations:  true,
		NotifyForConstraints: false,
		ImplicitSubscription: true,
		DependentOperations: []scdmodels.ID{
			scdmodels.ID("c09bcff5-35a4-41de-9220-6c140a9857ee"),
			scdmodels.ID("2cff1c62-cf1a-41ad-826b-d12dad432f21"),
		},
	}
	return result
}

// SearchSubscriptions is a stubbed implementation of SearchSubscriptions.
func (s *DummyStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*scdmodels.Subscription, error) {
	subs := []*scdmodels.Subscription{
		makeDummySubscription(scdmodels.ID("444eab15-8384-4e39-8589-5161689aee56")),
	}
	return subs, nil
}

// GetSubscription is a stubbed implementation of GetSubscription.
func (s *DummyStore) GetSubscription(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Subscription, error) {
	return makeDummySubscription(id), nil
}

// UpsertSubscription is a stubbed implementation of UpsertSubscription.
func (s *DummyStore) UpsertSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, []*scdmodels.Operation, error) {
	return sub, []*scdmodels.Operation{makeDummyOperation(sub.ID)}, nil
}

// DeleteSubscription is a stubbed implementation of DeleteSubscription.
func (s *DummyStore) DeleteSubscription(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner, version scdmodels.Version) (*scdmodels.Subscription, error) {
	sub := makeDummySubscription(id)
	sub.ID = id
	return sub, nil
}

// GetOperation is a stubbed implementation of GetOperation.
func (s *DummyStore) GetOperation(ctx context.Context, id scdmodels.ID) (*scdmodels.Operation, error) {
	return makeDummyOperation(id), nil
}

// DeleteOperation is a stubbed implementation of DeleteOperation.
func (s *DummyStore) DeleteOperation(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	return makeDummyOperation(id), []*scdmodels.Subscription{
		makeDummySubscription(scdmodels.ID("444eab15-8384-4e39-8589-5161689aee56")),
	}, nil
}

// UpsertOperation is a stubbed implementation of UpsertOperation.
func (s *DummyStore) UpsertOperation(ctx context.Context, operation *scdmodels.Operation, key []scdmodels.OVN) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	return operation, []*scdmodels.Subscription{
		makeDummySubscription(scdmodels.ID("444eab15-8384-4e39-8589-5161689aee56")),
	}, nil
}

// SearchOperations is a stubbed implementation of SearchOperations.
func (s *DummyStore) SearchOperations(ctx context.Context, v4d *dssmodels.Volume4D, owner dssmodels.Owner) ([]*scdmodels.Operation, error) {
	return []*scdmodels.Operation{
		makeDummyOperation("444eab15-8384-4e39-8589-5161689aee56"),
	}, nil
}
