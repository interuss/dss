package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

// OperationStore abstracts operation-specific interactions with the backing data store.
type OperationStore interface {
	// GetOperation returns the operation identified by "id".
	GetOperation(id scdmodels.ID) (*scdmodels.Operation, error)

	// DeleteOperation deletes the operation identified by "id" and owned by "owner".
	// Returns the deleted Operation and all Subscriptions affected by the delete.
	DeleteOperation(id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Operation, []*scdmodels.Subscription, error)

	// UpsertOperation inserts or updates an operation using key as a fencing
	// token. If operation does not reference an existing subscription, an
	// implicit subscription with parameters notifySubscriptionForConstraints
	// and subscriptionBaseURL is created.
	UpsertOperation(operation *scdmodels.Operation, key []scdmodels.OVN) (*scdmodels.Operation, []*scdmodels.Subscription, error)

	// SearchOperations returns all operations ownded by "owner" intersecting "v4d".
	SearchOperations(v4d *dssmodels.Volume4D, owner dssmodels.Owner) ([]*scdmodels.Operation, error)
}

// SubscriptionStore abstracts subscription-specific interactions with the backing data store.
type SubscriptionStore interface {
	// SearchSubscriptions returns all Subscriptions owned by "owner" in "cells".
	SearchSubscriptions(cells s2.CellUnion, owner dssmodels.Owner) ([]*scdmodels.Subscription, error)

	// GetSubscription returns the Subscription referenced by id, or nil if the
	// Subscription doesn't exist
	GetSubscription(id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Subscription, error)

	// UpsertSubscription upserts sub into the store and returns the result
	// subscription.
	UpsertSubscription(sub *scdmodels.Subscription) (*scdmodels.Subscription, []*scdmodels.Operation, error)

	// DeleteSubscription deletes a Subscription from the store and returns the
	// deleted subscription.  Returns nil and an error if the Subscription does
	// not exist, or is owned by someone other than the specified owner.
	DeleteSubscription(id scdmodels.ID, owner dssmodels.Owner, version scdmodels.Version) (*scdmodels.Subscription, error)
}

// Store abstracts strategic conflict detection interactions with the backing
// data store.
type Store interface {
	OperationStore
	SubscriptionStore
}

type Transaction interface {
	// Retrieve Store that operates within this Transaction.
	Store() (Store, error)

	// Commit commits all the operations performed on the Transactor so far.
	Commit() error

	// Rollback rolls back all the operations performed on the Transactor so far.
	Rollback() error
}

type Transactor interface {
	// Transact begins an atomic transaction
	Transact(ctx context.Context) (Transaction, error)
}

// TransactionOperation is an application action involving one or more chained
// Store actions joined by application logic.
type TransactionOperation func(store Store) (err error)

// PerformOperationWithRetries creates a Transaction from the Transactor,
// attempts to perform the provided action, and retries this process again if
// it fails in a retryable way.
func PerformOperationWithRetries(ctx context.Context, transactor Transactor, operation TransactionOperation, retries int) error {
	var err error
	var tx Transaction
	for i := 0; i <= retries; i++ {
		// Prepare a Store for `operation` to act on
		tx, err := transactor.Transact(ctx)
		if err != nil {
			return err
		}

		store, err := tx.Store()
		if err != nil {
			return err
		}

		err = operation(store)
		if err == nil {
			// Operation was successful
			err = tx.Commit()
			if err != nil && err != sql.ErrTxDone {
				// Commit errors are assumed to be retryable
				continue
			}
			// TransactionOperation and Commit were successful
			return nil
		}

		// A non-retryable error occurred
		rollbackErr := tx.Rollback()
		if rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			return errors.New(fmt.Sprintf(
				"error rolling back transaction after unsuccessful operation attempt: `%s` after `%s`",
				rollbackErr, err))
		}
		return err
	}

	// We've reached the maximum number of retries
	if tx != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			return errors.New(fmt.Sprintf(
				"error rolling back transaction after maximum retries: `%s` after `%s`",
				rollbackErr, err))
		}
	}
	return err
}
