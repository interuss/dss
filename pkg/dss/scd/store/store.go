package store

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/dss/models"
	scdmodels "github.com/interuss/dss/pkg/dss/scd/models"
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
	SearchOperations(ctx context.Context, v4d *dssmodels.Volume4D, owner dssmodels.Owner) ([]*scdmodels.Operation, error)
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
	UpsertSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error)

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

func MakeDummyOperation(id scdmodels.ID) *scdmodels.Operation {
	var (
		altLo = float32(11235)
		altHi = float32(81321)
		start = time.Now()
		end   = start.Add(2 * time.Minute)
	)

	return &scdmodels.Operation{
		ID:            id,
		Version:       314,
		OVN:           scdmodels.NewOVNFromTime(time.Now()),
		USSBaseURL:    "https://exampleuss.com/utm",
		AltitudeLower: &altLo,
		AltitudeUpper: &altHi,
		StartTime:     &start,
		EndTime:       &end,
		State:         scdmodels.OperationStateAccepted,
	}

}

// MakeDummySubscription returns a dummy subscription instance with ID id.
func MakeDummySubscription(id scdmodels.ID) *scdmodels.Subscription {
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

func (s *DummyStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*scdmodels.Subscription, error) {
	subs := []*scdmodels.Subscription{
		MakeDummySubscription(scdmodels.ID("444eab15-8384-4e39-8589-5161689aee56")),
	}
	return subs, nil
}

func (s *DummyStore) GetSubscription(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Subscription, error) {
	return MakeDummySubscription(id), nil
}

func (s *DummyStore) UpsertSubscription(ctx context.Context, sub *scdmodels.Subscription) (*scdmodels.Subscription, error) {
	return sub, nil
}

func (s *DummyStore) DeleteSubscription(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner, version scdmodels.Version) (*scdmodels.Subscription, error) {
	sub := MakeDummySubscription(id)
	sub.ID = id
	return sub, nil
}

func (s *DummyStore) GetOperation(ctx context.Context, id scdmodels.ID) (*scdmodels.Operation, error) {
	return MakeDummyOperation(id), nil
}

func (s *DummyStore) DeleteOperation(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	return MakeDummyOperation(id), []*scdmodels.Subscription{
		MakeDummySubscription(scdmodels.ID("444eab15-8384-4e39-8589-5161689aee56")),
	}, nil
}

func (s *DummyStore) UpsertOperation(ctx context.Context, operation *scdmodels.Operation, key []scdmodels.OVN) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	return operation, []*scdmodels.Subscription{
		MakeDummySubscription(scdmodels.ID("444eab15-8384-4e39-8589-5161689aee56")),
	}, nil
}

func (s *DummyStore) SearchOperations(ctx context.Context, v4d *dssmodels.Volume4D, owner dssmodels.Owner) ([]*scdmodels.Operation, error) {
	return []*scdmodels.Operation{
		MakeDummyOperation("444eab15-8384-4e39-8589-5161689aee56"),
	}, nil
}
