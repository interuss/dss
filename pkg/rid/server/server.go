package server

import (
	"time"

	"github.com/interuss/dss/pkg/auth"
	"github.com/interuss/dss/pkg/rid/application"
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

// Server implements ridpb.DiscoveryAndSynchronizationService.
type Server struct {
	App     application.App
	Timeout time.Duration
}

// AuthScopes returns a map of endpoint to required Oauth scope.
func AuthScopes() map[auth.Operation][]auth.Scope {
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
