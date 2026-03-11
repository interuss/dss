// This file is auto-generated; do not change as any changes will be overwritten
package auxv1

import (
	"context"
	"fmt"
	"github.com/interuss/dss/pkg/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"regexp"
)

type APIRouter struct {
	Routes         []*api.Route
	Implementation Implementation
	Authorizer     api.Authorizer
}

var tracer = otel.Tracer("auxv1.api")

// *auxv1.APIRouter (type defined above) implements the api.PartialRouter interface
func (s *APIRouter) Handle(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range s.Routes {
		if route.Method == r.Method && route.Pattern.MatchString(r.URL.Path) {

			span := trace.SpanFromContext(r.Context())

			if span.IsRecording() {
				// Current span is the one from the otelhttp handler
				span.SetName(fmt.Sprintf("%s %s", r.Method, route.Path))
			}

			ctx, span := tracer.Start(r.Context(), route.Name)
			defer span.End()
			r = r.WithContext(ctx)

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
	if query.Has("owner") {
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

func (s *APIRouter) GetPool(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetPoolRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetPoolSecurity)

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response := s.Implementation.GetPool(ctx, &req)

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
	if response.Response501 != nil {
		api.WriteJSON(w, 501, response.Response501)
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
	if response.Response501 != nil {
		api.WriteJSON(w, 501, response.Response501)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) PutDSSInstancesHeartbeat(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req PutDSSInstancesHeartbeatRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, PutDSSInstancesHeartbeatSecurity)

	// Copy query parameters
	query := r.URL.Query()
	if query.Has("source") {
		v := query.Get("source")
		req.Source = &v
	}
	if query.Has("timestamp") {
		v := query.Get("timestamp")
		req.Timestamp = &v
	}
	if query.Has("next_heartbeat_expected_before") {
		v := query.Get("next_heartbeat_expected_before")
		req.NextHeartbeatExpectedBefore = &v
	}

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response := s.Implementation.PutDSSInstancesHeartbeat(ctx, &req)

	// Write response to client
	if response.Response201 != nil {
		api.WriteJSON(w, 201, response.Response201)
		return
	}
	if response.Response400 != nil {
		api.WriteJSON(w, 400, response.Response400)
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
	if response.Response501 != nil {
		api.WriteJSON(w, 501, response.Response501)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) GetAcceptedCAs(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetAcceptedCAsRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetAcceptedCAsSecurity)

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response := s.Implementation.GetAcceptedCAs(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response501 != nil {
		api.WriteJSON(w, 501, response.Response501)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) GetInstanceCAs(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetInstanceCAsRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetInstanceCAsSecurity)

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response := s.Implementation.GetInstanceCAs(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response501 != nil {
		api.WriteJSON(w, 501, response.Response501)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) GetScdLockMode(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetScdLockModeRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetScdLockModeSecurity)

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response := s.Implementation.GetScdLockMode(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response501 != nil {
		api.WriteJSON(w, 501, response.Response501)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func MakeAPIRouter(impl Implementation, auth api.Authorizer) APIRouter {
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 8)}

	pattern := regexp.MustCompile("^/aux/v1/version$")
	router.Routes[0] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetVersion, Name: "auxv1.GetVersion", Path: "/aux/v1/version"}

	pattern = regexp.MustCompile("^/aux/v1/validate_oauth$")
	router.Routes[1] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.ValidateOauth, Name: "auxv1.ValidateOauth", Path: "/aux/v1/validate_oauth"}

	pattern = regexp.MustCompile("^/aux/v1/pool$")
	router.Routes[2] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetPool, Name: "auxv1.GetPool", Path: "/aux/v1/pool"}

	pattern = regexp.MustCompile("^/aux/v1/pool/dss_instances$")
	router.Routes[3] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetDSSInstances, Name: "auxv1.GetDSSInstances", Path: "/aux/v1/pool/dss_instances"}

	pattern = regexp.MustCompile("^/aux/v1/pool/dss_instances/heartbeat$")
	router.Routes[4] = &api.Route{Method: http.MethodPut, Pattern: pattern, Handler: router.PutDSSInstancesHeartbeat, Name: "auxv1.PutDSSInstancesHeartbeat", Path: "/aux/v1/pool/dss_instances/heartbeat"}

	pattern = regexp.MustCompile("^/aux/v1/configuration/accepted_ca_certs$")
	router.Routes[5] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetAcceptedCAs, Name: "auxv1.GetAcceptedCAs", Path: "/aux/v1/configuration/accepted_ca_certs"}

	pattern = regexp.MustCompile("^/aux/v1/configuration/ca_certs$")
	router.Routes[6] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetInstanceCAs, Name: "auxv1.GetInstanceCAs", Path: "/aux/v1/configuration/ca_certs"}

	pattern = regexp.MustCompile("^/aux/v1/configuration/scd_lock_mode$")
	router.Routes[7] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetScdLockMode, Name: "auxv1.GetScdLockMode", Path: "/aux/v1/configuration/scd_lock_mode"}

	return router
}
