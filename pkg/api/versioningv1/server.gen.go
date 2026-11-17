// This file is auto-generated; do not change as any changes will be overwritten
package versioningv1

import (
	"context"
	"fmt"
	"github.com/interuss/dss/pkg/api"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"regexp"
)

type APIRouter struct {
	Routes         []*api.Route
	Implementation Implementation
	Authorizer     api.Authorizer
}

var tracer = otel.Tracer("versioningv1.api")

// *versioningv1.APIRouter (type defined above) implements the api.PartialRouter interface
func (s *APIRouter) Handle(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range s.Routes {
		if route.Method == r.Method && route.Pattern.MatchString(r.URL.Path) {

			if labeler, ok := otelhttp.LabelerFromContext(r.Context()); ok {
				labeler.Add(semconv.HTTPRoute(route.Path))
			}

			// We retrieve the current span from the otelhttp handler to set its name property.
			span := trace.SpanFromContext(r.Context())

			if span.IsRecording() { // If the span is not recording, the name cannot be changed. This also likely means the otelhttp handler is not present (tracing disabled).
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
	var response GetVersionResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.SystemIdentity = SystemBoundaryIdentifier(pathMatch[1])

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetVersionSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.GetVersion(ctx, &req)
	}

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
	if response.Response404 != nil {
		api.WriteJSON(w, 404, response.Response404)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func setAuthError(ctx context.Context, authErr error, resp401 **api.ClientErrorBody, resp403 **api.ClientErrorBody, resp500 **api.InternalServerErrorBody) {
	switch stacktrace.GetCode(authErr) {
	case dsserr.Unauthenticated:
		*resp401 = &api.ClientErrorBody{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authentication failed"))}
	case dsserr.PermissionDenied:
		*resp403 = &api.ClientErrorBody{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authorization failed"))}
	default:
		if authErr == nil {
			authErr = stacktrace.NewError("Unknown error")
		}
		*resp500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Could not perform authorization"))}
	}
}

func MakeAPIRouter(impl Implementation, auth api.Authorizer) APIRouter {
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 1)}

	pattern := regexp.MustCompile("^/versions/(?P<system_identity>[^/]*)$")
	router.Routes[0] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetVersion, Name: "versioningv1.GetVersion", Path: "/versions/{system_identity}"}

	return router
}
