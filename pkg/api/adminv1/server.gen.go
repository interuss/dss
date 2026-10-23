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

func (s *APIRouter) GetThisRaftNode(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetThisRaftNodeRequest
	var response GetThisRaftNodeResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.GroupName = RaftGroupName(pathMatch[1])

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.GetThisRaftNode(ctx, &req)

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

func (s *APIRouter) CreateRaftGroup(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req CreateRaftGroupRequest
	var response CreateRaftGroupResponseSet

	// Parse request body
	req.Body = new(CreateRaftGroupRequestBody)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.CreateRaftGroup(ctx, &req)

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
	if response.Response409 != nil {
		api.WriteJSON(w, 409, response.Response409)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) GetRaftGroup(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetRaftGroupRequest
	var response GetRaftGroupResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.GroupName = RaftGroupName(pathMatch[1])

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.GetRaftGroup(ctx, &req)

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

func (s *APIRouter) CreateRaftMember(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req CreateRaftMemberRequest
	var response CreateRaftMemberResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.GroupName = RaftGroupName(pathMatch[1])

	// Parse request body
	req.Body = new(CreateRaftMemberRequestBody)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.CreateRaftMember(ctx, &req)

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

func (s *APIRouter) DeleteRaftMember(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req DeleteRaftMemberRequest
	var response DeleteRaftMemberResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.GroupName = RaftGroupName(pathMatch[1])
	req.NodeId = pathMatch[2]

	// Call implementation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	response = s.Implementation.DeleteRaftMember(ctx, &req)

	// Write response to client
	if response.Response204 != nil {
		api.WriteJSON(w, 204, response.Response204)
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
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 5)}

	pattern := regexp.MustCompile("^/admin/v1/raft/groups/(?P<group_name>[^/]*)/this_node$")
	router.Routes[0] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetThisRaftNode, Name: "adminv1.GetThisRaftNode", Path: "/admin/v1/raft/groups/{group_name}/this_node"}

	pattern = regexp.MustCompile("^/admin/v1/raft/groups$")
	router.Routes[1] = &api.Route{Method: http.MethodPost, Pattern: pattern, Handler: router.CreateRaftGroup, Name: "adminv1.CreateRaftGroup", Path: "/admin/v1/raft/groups"}

	pattern = regexp.MustCompile("^/admin/v1/raft/groups/(?P<group_name>[^/]*)$")
	router.Routes[2] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetRaftGroup, Name: "adminv1.GetRaftGroup", Path: "/admin/v1/raft/groups/{group_name}"}

	pattern = regexp.MustCompile("^/admin/v1/raft/groups/(?P<group_name>[^/]*)/members$")
	router.Routes[3] = &api.Route{Method: http.MethodPost, Pattern: pattern, Handler: router.CreateRaftMember, Name: "adminv1.CreateRaftMember", Path: "/admin/v1/raft/groups/{group_name}/members"}

	pattern = regexp.MustCompile("^/admin/v1/raft/groups/(?P<group_name>[^/]*)/members/(?P<node_id>[^/]*)$")
	router.Routes[4] = &api.Route{Method: http.MethodDelete, Pattern: pattern, Handler: router.DeleteRaftMember, Name: "adminv1.DeleteRaftMember", Path: "/admin/v1/raft/groups/{group_name}/members/{node_id}"}

	return router
}
