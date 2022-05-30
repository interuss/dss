package v1

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
		"/ridpb.DiscoveryAndSynchronizationService/CreateIdentificationServiceArea":  auth.RequireAllScopes(Scopes.ISA.Write),
		"/ridpb.DiscoveryAndSynchronizationService/DeleteIdentificationServiceArea":  auth.RequireAllScopes(Scopes.ISA.Write),
		"/ridpb.DiscoveryAndSynchronizationService/GetIdentificationServiceArea":     auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpb.DiscoveryAndSynchronizationService/SearchIdentificationServiceAreas": auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpb.DiscoveryAndSynchronizationService/UpdateIdentificationServiceArea":  auth.RequireAllScopes(Scopes.ISA.Write),
		"/ridpb.DiscoveryAndSynchronizationService/CreateSubscription":               auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpb.DiscoveryAndSynchronizationService/DeleteSubscription":               auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpb.DiscoveryAndSynchronizationService/GetSubscription":                  auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpb.DiscoveryAndSynchronizationService/SearchSubscriptions":              auth.RequireAllScopes(Scopes.ISA.Read),
		"/ridpb.DiscoveryAndSynchronizationService/UpdateSubscription":               auth.RequireAllScopes(Scopes.ISA.Read),
	}
}
