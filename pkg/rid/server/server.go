package server

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"

	"github.com/interuss/dss/pkg/auth"
)

var (
	// Scopes bundles up auth scopes for the remote-id server.
	Scopes = struct {
		ISA struct {
			Write auth.Scope
			Read  auth.Scope
		}
	}{
		ISA: struct {
			Write auth.Scope
			Read  auth.Scope
		}{
			Write: "dss.write.identification_service_areas",
			Read:  "dss.read.identification_service_areas",
		},
	}
)

// Store provides an interface for storing DSS data.
type Store interface {
	// Close closes the store and should release all resources.
	Close() error

	GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error)

	// Delete deletes the IdentificationServiceArea identified by "id" and owned by "owner".
	// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
	DeleteISA(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error)

	// InsertISA inserts or updates an ISA.
	InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error)

	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error)

	// GetSubscription returns the subscription identified by "id".
	GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error)

	// Delete deletes the subscription identified by "id" and
	// returns the deleted subscription.
	DeleteSubscription(ctx context.Context, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ridmodels.Subscription, error)

	// InsertSubscription inserts or updates a subscription.
	InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error)

	// SearchSubscriptions returns all subscriptions ownded by "owner" in "cells".
	SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error)
}

// NewNilStore returns a nil Store instance.
func NewNilStore() Store {
	return nil
}

// Server implements ridpb.DiscoveryAndSynchronizationService.
type Server struct {
	Store   Store
	Timeout time.Duration
}

// AuthScopes returns a map of endpoint to required Oauth scope.
func (s *Server) AuthScopes() map[auth.Operation][]auth.Scope {
	return map[auth.Operation][]auth.Scope{
		"/ridpb.DiscoveryAndSynchronizationService/CreateIdentificationServiceArea":  {Scopes.ISA.Write},
		"/ridpb.DiscoveryAndSynchronizationService/DeleteIdentificationServiceArea":  {Scopes.ISA.Write},
		"/ridpb.DiscoveryAndSynchronizationService/GetIdentificationServiceArea":     {Scopes.ISA.Read},
		"/ridpb.DiscoveryAndSynchronizationService/SearchIdentificationServiceAreas": {Scopes.ISA.Read},
		"/ridpb.DiscoveryAndSynchronizationService/UpdateIdentificationServiceArea":  {Scopes.ISA.Write},
		"/ridpb.DiscoveryAndSynchronizationService/CreateSubscription":               {Scopes.ISA.Write},
		"/ridpb.DiscoveryAndSynchronizationService/DeleteSubscription":               {Scopes.ISA.Write},
		"/ridpb.DiscoveryAndSynchronizationService/GetSubscription":                  {Scopes.ISA.Read},
		"/ridpb.DiscoveryAndSynchronizationService/SearchSubscriptions":              {Scopes.ISA.Read},
		"/ridpb.DiscoveryAndSynchronizationService/UpdateSubscription":               {Scopes.ISA.Write},
	}
}
