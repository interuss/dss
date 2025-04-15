// This file is auto-generated; do not change as any changes will be overwritten
package auxv1

import (
	"context"
	"github.com/interuss/dss/pkg/api"
	"net/http"
	"regexp"
)

type APIRouter struct {
	Routes         []*api.Route
	Implementation Implementation
	Authorizer     api.Authorizer
}

// *auxv1.APIRouter (type defined above) implements the api.PartialRouter interface
func (s *APIRouter) Handle(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range s.Routes {
		if route.Method == r.Method && route.Pattern.MatchString(r.URL.Path) {
			route.Handler(route.Pattern, w, r)
			return true
		}
	}
	return false
}

func (s *APIRouter) GetVersion(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetVersionRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetVersionSecurity)

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response := s.Implementation.GetVersion(ctx, &req)

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

func (s *APIRouter) ValidateOauth(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req ValidateOauthRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, ValidateOauthSecurity)

	// Copy query parameters
	query := r.URL.Query()
	// TODO: Change to query.Has after Go 1.17
	if query.Get("owner") != "" {
		v := query.Get("owner")
		req.Owner = &v
	}

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response := s.Implementation.ValidateOauth(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response401 != nil {
		api.WriteJSON(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		api.WriteJSON(w, 403, response.Response403)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) GetDSSInstances(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetDSSInstancesRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetDSSInstancesSecurity)

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response := s.Implementation.GetDSSInstances(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response401 != nil {
		api.WriteJSON(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		api.WriteJSON(w, 403, response.Response403)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func MakeAPIRouter(impl Implementation, auth api.Authorizer) APIRouter {
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 3)}

	pattern := regexp.MustCompile("^/aux/v1/version$")
	router.Routes[0] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetVersion}

	pattern = regexp.MustCompile("^/aux/v1/validate_oauth$")
	router.Routes[1] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.ValidateOauth}

	pattern = regexp.MustCompile("^/aux/v1/pool/dss_instances$")
	router.Routes[2] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetDSSInstances}

	return router
}
