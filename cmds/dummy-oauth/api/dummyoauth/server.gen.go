// This file is auto-generated; do not change as any changes will be overwritten
package dummyoauth

import (
	"context"
	"github.com/interuss/dss/cmds/dummy-oauth/api"
	"go.opentelemetry.io/otel"
	"net/http"
	"regexp"
	"strconv"
)

type APIRouter struct {
	Routes         []*api.Route
	Implementation Implementation
	Authorizer     api.Authorizer
}

var tracer = otel.Tracer("dummyoauth.api")

// *dummyoauth.APIRouter (type defined above) implements the api.PartialRouter interface
func (s *APIRouter) MatchRoute(r *http.Request) (bool, *api.Route) {
	for _, route := range s.Routes {
		if route.Method == r.Method && route.Pattern.MatchString(r.URL.Path) {
			return true, route
		}
	}
	return false, nil
}

func (s *APIRouter) GetToken(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetTokenRequest
	var response GetTokenResponseSet

	// Copy query parameters
	query := r.URL.Query()
	if query.Has("intended_audience") {
		v := query.Get("intended_audience")
		req.IntendedAudience = &v
	}
	if query.Has("scope") {
		v := query.Get("scope")
		req.Scope = &v
	}
	if query.Has("issuer") {
		v := query.Get("issuer")
		req.Issuer = &v
	}
	if query.Has("expire") {
		i, err := strconv.ParseInt(query.Get("expire"), 10, 64)
		if err == nil {
			req.Expire = &i
		}
	}
	if query.Has("sub") {
		v := query.Get("sub")
		req.Sub = &v
	}

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.GetToken(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		api.WriteJSON(w, 400, response.Response400)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) GetJWKS(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetJWKSRequest
	var response GetJWKSResponseSet

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.GetJWKS(ctx, &req)

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
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 2)}

	pattern := regexp.MustCompile("^/token$")
	router.Routes[0] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetToken, Name: "dummyoauth.GetToken", Path: "/token", Tracer: tracer}

	pattern = regexp.MustCompile("^/.well-known/jwks.json$")
	router.Routes[1] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetJWKS, Name: "dummyoauth.GetJWKS", Path: "/.well-known/jwks.json", Tracer: tracer}

	return &router
}
