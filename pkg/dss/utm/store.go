package utm

import (
	"context"

	"github.com/golang/geo/s2"
	commonmodels "github.com/interuss/dss/pkg/dss/models"
	"github.com/interuss/dss/pkg/dss/utm/models"
)

// Store abstracts interactions with a backing data store.
type Store interface {
	// SearchSubscriptions returns all Subscriptions owned by "owner" in "cells".
	SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner commonmodels.Owner) ([]*models.Subscription, error)

	// GetSubscription returns the Subscription referenced by id, or nil if the
	// Subscription doesn't exist
	GetSubscription(ctx context.Context, id models.ID, owner commonmodels.Owner) (*models.Subscription, error)

	// InsertSubscription inserts sub into the store and returns the result
	// subscription.
	InsertSubscription(ctx context.Context, sub *models.Subscription, owner commonmodels.Owner) (*models.Subscription, error)

	// DeleteSubscription deletes a Subscription from the store and returns the
	// deleted subscription.  Returns nil and an error if the Subscription does
	// not exist, or is owned by someone other than the specified owner.
	DeleteSubscription(ctx context.Context, id models.ID, owner commonmodels.Owner) (*models.Subscription, error)
}

// DummyStore implements Store interface entirely with error-free no-ops
type DummyStore struct {
}

// MakeDummySubscription returns a dummy subscription instance with ID id.
func MakeDummySubscription(id models.ID) *models.Subscription {
	altLo := float32(11235)
	altHi := float32(81321)
	result := &models.Subscription{
		ID:                   id,
		Version:              314,
		NotificationIndex:    123,
		BaseURL:              "https://exampleuss.com/utm",
		AltitudeLo:           &altLo,
		AltitudeHi:           &altHi,
		NotifyForOperations:  true,
		NotifyForConstraints: false,
		ImplicitSubscription: true,
		DependentOperations: []models.ID{
			models.ID("c09bcff5-35a4-41de-9220-6c140a9857ee"),
			models.ID("2cff1c62-cf1a-41ad-826b-d12dad432f21"),
		},
	}
	return result
}

func (s *DummyStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner commonmodels.Owner) ([]*models.Subscription, error) {
	subs := []*models.Subscription{
		MakeDummySubscription(models.ID("444eab15-8384-4e39-8589-5161689aee56")),
	}
	return subs, nil
}

func (s *DummyStore) GetSubscription(ctx context.Context, id models.ID, owner commonmodels.Owner) (*models.Subscription, error) {
	return MakeDummySubscription(id), nil
}

func (s *DummyStore) InsertSubscription(ctx context.Context, sub *models.Subscription, owner commonmodels.Owner) (*models.Subscription, error) {
	return sub, nil
}

func (s *DummyStore) DeleteSubscription(ctx context.Context, id models.ID, owner commonmodels.Owner) (*models.Subscription, error) {
	sub := MakeDummySubscription(id)
	sub.ID = id
	return sub, nil
}
