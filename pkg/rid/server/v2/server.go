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
		ServiceProvider auth.Scope
		DisplayProvider auth.Scope
	}{
		ServiceProvider: "rid.server_provider",
		DisplayProvider: "rid.display_provider",
	}
)

// Server implements ridpbv2.StandardRemoteIDAPIInterfacesServiceServer.
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
		"/ridpbv2.StandardRemoteIDAPIInterfacesService/CreateIdentificationServiceArea":  auth.RequireAllScopes(Scopes.ServiceProvider),
		"/ridpbv2.StandardRemoteIDAPIInterfacesService/DeleteIdentificationServiceArea":  auth.RequireAllScopes(Scopes.ServiceProvider),
		"/ridpbv2.StandardRemoteIDAPIInterfacesService/GetIdentificationServiceArea":     auth.RequireAnyScope(Scopes.ServiceProvider, Scopes.DisplayProvider),
		"/ridpbv2.StandardRemoteIDAPIInterfacesService/SearchIdentificationServiceAreas": auth.RequireAnyScope(Scopes.ServiceProvider, Scopes.DisplayProvider),
		"/ridpbv2.StandardRemoteIDAPIInterfacesService/UpdateIdentificationServiceArea":  auth.RequireAllScopes(Scopes.ServiceProvider),
		"/ridpbv2.StandardRemoteIDAPIInterfacesService/CreateSubscription":               auth.RequireAllScopes(Scopes.DisplayProvider),
		"/ridpbv2.StandardRemoteIDAPIInterfacesService/DeleteSubscription":               auth.RequireAllScopes(Scopes.DisplayProvider),
		"/ridpbv2.StandardRemoteIDAPIInterfacesService/GetSubscription":                  auth.RequireAllScopes(Scopes.DisplayProvider),
		"/ridpbv2.StandardRemoteIDAPIInterfacesService/SearchSubscriptions":              auth.RequireAllScopes(Scopes.DisplayProvider),
		"/ridpbv2.StandardRemoteIDAPIInterfacesService/UpdateSubscription":               auth.RequireAllScopes(Scopes.DisplayProvider),
	}
}
