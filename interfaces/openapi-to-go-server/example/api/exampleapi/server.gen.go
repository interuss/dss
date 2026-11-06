// This file is auto-generated; do not change as any changes will be overwritten
package exampleapi

import (
	"context"
	"example/api"
	"fmt"
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

var tracer = otel.Tracer("exampleapi.api")

// *exampleapi.APIRouter (type defined above) implements the api.PartialRouter interface
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

func MakeAPIRouter(impl Implementation, auth api.Authorizer) APIRouter {
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 1)}

	pattern := regexp.MustCompile("^/exampleapi/greeting$")
	router.Routes[0] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetGreeting, Name: "exampleapi.GetGreeting", Path: "/exampleapi/greeting"}

	return router
}
