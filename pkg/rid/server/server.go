package server

import (
	"time"

	"github.com/robfig/cron/v3"

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
	App        application.App
	Timeout    time.Duration
	Locality   string
	EnableHTTP bool
	Cron       *cron.Cron
}

// AuthScopes returns a map of endpoint to required Oauth scope.
func (s *Server) AuthScopes() map[auth.Operation]auth.KeyClaimedScopesValidator {
	return map[auth.Operation]auth.KeyClaimedScopesValidator{
		"/ridpbv1.DiscoveryAndSynchronizationService/CreateIdentificationServiceArea":  auth.RequireAllScopes(Scopes.ISA.Write),
		"/ridpbv1.DiscoveryAndSynchronizationService/DeleteIdentificationServiceArea":  auth.RequireAllScopes(Scopes.ISA.Write),
		"/ridpbv1.DiscoveryAndSynchronizationService/GetIdentificationServiceArea":     auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpbv1.DiscoveryAndSynchronizationService/SearchIdentificationServiceAreas": auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpbv1.DiscoveryAndSynchronizationService/UpdateIdentificationServiceArea":  auth.RequireAllScopes(Scopes.ISA.Write),
		"/ridpbv1.DiscoveryAndSynchronizationService/CreateSubscription":               auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpbv1.DiscoveryAndSynchronizationService/DeleteSubscription":               auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpbv1.DiscoveryAndSynchronizationService/GetSubscription":                  auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpbv1.DiscoveryAndSynchronizationService/SearchSubscriptions":              auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpbv1.DiscoveryAndSynchronizationService/UpdateSubscription":               auth.RequireAllScopes(Scopes.ISA.Read),
	}
}
