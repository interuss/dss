// This file is auto-generated; do not change as any changes will be overwritten
package adminv1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/interuss/dss/pkg/api"
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

var tracer = otel.Tracer("adminv1.api")

// *adminv1.APIRouter (type defined above) implements the api.PartialRouter interface
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

func (s *APIRouter) GetRaftMembers(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetRaftMembersRequest
	var response GetRaftMembersResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Group = pathMatch[1]

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.GetRaftMembers(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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

func (s *APIRouter) AddRaftMember(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req AddRaftMemberRequest
	var response AddRaftMemberResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Group = pathMatch[1]

	// Parse request body
	req.Body = new(RaftMember)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.AddRaftMember(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		api.WriteJSON(w, 400, response.Response400)
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

func (s *APIRouter) RemoveRaftMember(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req RemoveRaftMemberRequest
	var response RemoveRaftMemberResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Group = pathMatch[1]
	req.NodeId = pathMatch[2]

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.RemoveRaftMember(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		api.WriteJSON(w, 400, response.Response400)
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

func MakeAPIRouter(impl Implementation, auth api.Authorizer) APIRouter {
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 3)}

	pattern := regexp.MustCompile("^/admin/v1/raft/(?P<group>[^/]*)/members$")
	router.Routes[0] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetRaftMembers, Name: "adminv1.GetRaftMembers", Path: "/admin/v1/raft/{group}/members"}

	pattern = regexp.MustCompile("^/admin/v1/raft/(?P<group>[^/]*)/members$")
	router.Routes[1] = &api.Route{Method: http.MethodPost, Pattern: pattern, Handler: router.AddRaftMember, Name: "adminv1.AddRaftMember", Path: "/admin/v1/raft/{group}/members"}

	pattern = regexp.MustCompile("^/admin/v1/raft/(?P<group>[^/]*)/members/(?P<node_id>[^/]*)$")
	router.Routes[2] = &api.Route{Method: http.MethodDelete, Pattern: pattern, Handler: router.RemoveRaftMember, Name: "adminv1.RemoveRaftMember", Path: "/admin/v1/raft/{group}/members/{node_id}"}

	return router
}
