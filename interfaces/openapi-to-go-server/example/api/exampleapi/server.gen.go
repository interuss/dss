// This file is auto-generated; do not change as any changes will be overwritten
package exampleapi

import (
	"context"
	"example/api"
	"go.opentelemetry.io/otel"
	"net/http"
	"regexp"
)

type APIRouter struct {
	Routes         []*api.Route
	Implementation Implementation
	Authorizer     api.Authorizer
}

var tracer = otel.Tracer("exampleapi.api")

// *exampleapi.APIRouter (type defined above) implements the api.PartialRouter interface
func (s *APIRouter) MatchRoute(r *http.Request) (bool, *api.Route) {
	for _, route := range s.Routes {
		if route.Method == r.Method && route.Pattern.MatchString(r.URL.Path) {
			return true, route
		}
	}
	return false, nil
}

func (s *APIRouter) GetGreeting(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetGreetingRequest
	var response GetGreetingResponseSet

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.GetGreeting(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func MakeAPIRouter(impl Implementation, auth api.Authorizer) *APIRouter {
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 1)}

	pattern := regexp.MustCompile("^/exampleapi/greeting$")
	router.Routes[0] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetGreeting, Name: "exampleapi.GetGreeting", Path: "/exampleapi/greeting", Tracer: tracer}

	return &router
}
