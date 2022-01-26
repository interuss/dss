// This file is auto-generated; do not change as any changes will be overwritten
package dummy_oauth

import (
	"context"
	"github.com/interuss/dss/cmds/dummy-oauth/api"
	"net/http"
	"regexp"
	"strconv"
)

type APIRouter struct {
	Routes         []*api.Route
	Implementation Implementation
	Authorizer     api.Authorizer
}

// *dummy_oauth.APIRouter (type defined above) implements the api.APIRouter interface
func (a *APIRouter) Handle(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range a.Routes {
		if route.Pattern.MatchString(r.URL.Path) {
			route.Handler(route.Pattern, w, r)
			return true
		}
	}
	return false
}

func (s *APIRouter) GetToken(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetTokenRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &GetTokenSecurity)

	// Copy query parameters
	query := r.URL.Query()
	// TODO: Change to query.Has after Go 1.17
	if query.Get("intended_audience") != "" {
		v := query.Get("intended_audience")
		req.IntendedAudience = &v
	}
	if query.Get("scope") != "" {
		v := query.Get("scope")
		req.Scope = &v
	}
	if query.Get("issuer") != "" {
		v := query.Get("issuer")
		req.Issuer = &v
	}
	if query.Get("expire") != "" {
		i, err := strconv.ParseInt(query.Get("expire"), 10, 64)
		if err == nil {
			req.Expire = &i
		}
	}
	if query.Get("sub") != "" {
		v := query.Get("sub")
		req.Sub = &v
	}

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.GetToken(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJson(w, 200, response.Response200)
		return
	}
	if response.Response500 != nil {
		api.WriteJson(w, 500, response.Response500)
		return
	}
	api.WriteJson(w, 500, api.InternalServerErrorBody{"Handler implementation did not set a response"})
}

func MakeAPIRouter(impl Implementation, auth api.Authorizer) APIRouter {
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 1)}

	pattern := regexp.MustCompile("^/token$")
	router.Routes[0] = &api.Route{Pattern: pattern, Handler: router.GetToken}

	return router
}
