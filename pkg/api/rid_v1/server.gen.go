// This file is auto-generated; do not change as any changes will be overwritten
package rid_v1

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/interuss/dss/pkg/api"
)

type APIRouter struct {
	Routes         []*api.Route
	Implementation Implementation
	Authorizer     api.Authorizer
}

// *rid_v1.APIRouter (type defined above) implements the api.PartialRouter interface
func (s *APIRouter) Handle(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range s.Routes {
		if route.Pattern.MatchString(r.URL.Path) {
			route.Handler(route.Pattern, w, r)
			return true
		}
	}
	return false
}

func (s *APIRouter) SearchIdentificationServiceAreas(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req SearchIdentificationServiceAreasRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &SearchIdentificationServiceAreasSecurity)

	// Copy query parameters
	query := r.URL.Query()
	// TODO: Change to query.Has after Go 1.17
	if query.Get("area") != "" {
		v := GeoPolygonString(query.Get("area"))
		req.Area = &v
	}
	if query.Get("earliest_time") != "" {
		v := query.Get("earliest_time")
		req.EarliestTime = &v
	}
	if query.Get("latest_time") != "" {
		v := query.Get("latest_time")
		req.LatestTime = &v
	}

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.SearchIdentificationServiceAreas(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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
	if response.Response413 != nil {
		api.WriteJSON(w, 413, response.Response413)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) GetIdentificationServiceArea(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetIdentificationServiceAreaRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &GetIdentificationServiceAreaSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Id = EntityUUID(pathMatch[1])

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.GetIdentificationServiceArea(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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

func (s *APIRouter) CreateIdentificationServiceArea(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req CreateIdentificationServiceAreaRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &CreateIdentificationServiceAreaSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Id = EntityUUID(pathMatch[1])

	// Parse request body
	req.Body = new(CreateIdentificationServiceAreaParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.CreateIdentificationServiceArea(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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
	if response.Response409 != nil {
		api.WriteJSON(w, 409, response.Response409)
		return
	}
	if response.Response413 != nil {
		api.WriteJSON(w, 413, response.Response413)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) UpdateIdentificationServiceArea(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req UpdateIdentificationServiceAreaRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &UpdateIdentificationServiceAreaSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Id = EntityUUID(pathMatch[1])
	req.Version = pathMatch[2]

	// Parse request body
	req.Body = new(UpdateIdentificationServiceAreaParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.UpdateIdentificationServiceArea(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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
	if response.Response409 != nil {
		api.WriteJSON(w, 409, response.Response409)
		return
	}
	if response.Response413 != nil {
		api.WriteJSON(w, 413, response.Response413)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) DeleteIdentificationServiceArea(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req DeleteIdentificationServiceAreaRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &DeleteIdentificationServiceAreaSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Id = EntityUUID(pathMatch[1])
	req.Version = pathMatch[2]

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.DeleteIdentificationServiceArea(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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

func (s *APIRouter) SearchSubscriptions(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req SearchSubscriptionsRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &SearchSubscriptionsSecurity)

	// Copy query parameters
	query := r.URL.Query()
	// TODO: Change to query.Has after Go 1.17
	if query.Get("area") != "" {
		v := GeoPolygonString(query.Get("area"))
		req.Area = &v
	}

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.SearchSubscriptions(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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
	if response.Response413 != nil {
		api.WriteJSON(w, 413, response.Response413)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) GetSubscription(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetSubscriptionRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &GetSubscriptionSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Id = SubscriptionUUID(pathMatch[1])

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.GetSubscription(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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

func (s *APIRouter) CreateSubscription(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &CreateSubscriptionSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Id = SubscriptionUUID(pathMatch[1])

	// Parse request body
	req.Body = new(CreateSubscriptionParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.CreateSubscription(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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
	if response.Response409 != nil {
		api.WriteJSON(w, 409, response.Response409)
		return
	}
	if response.Response429 != nil {
		api.WriteJSON(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) UpdateSubscription(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req UpdateSubscriptionRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &UpdateSubscriptionSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Id = SubscriptionUUID(pathMatch[1])
	req.Version = pathMatch[2]

	// Parse request body
	req.Body = new(UpdateSubscriptionParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.UpdateSubscription(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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
	if response.Response409 != nil {
		api.WriteJSON(w, 409, response.Response409)
		return
	}
	if response.Response429 != nil {
		api.WriteJSON(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) DeleteSubscription(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req DeleteSubscriptionRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &DeleteSubscriptionSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Id = SubscriptionUUID(pathMatch[1])
	req.Version = pathMatch[2]

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.DeleteSubscription(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
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

func MakeAPIRouter(impl Implementation, auth api.Authorizer) APIRouter {
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 10)}

	pattern := regexp.MustCompile("^/rid_v1/v1/dss/identification_service_areas$")
	router.Routes[0] = &api.Route{Pattern: pattern, Handler: router.SearchIdentificationServiceAreas}

	pattern = regexp.MustCompile("^/rid_v1/v1/dss/identification_service_areas/(?P<id>[^/]*)$")
	router.Routes[1] = &api.Route{Pattern: pattern, Handler: router.GetIdentificationServiceArea}

	pattern = regexp.MustCompile("^/rid_v1/v1/dss/identification_service_areas/(?P<id>[^/]*)$")
	router.Routes[2] = &api.Route{Pattern: pattern, Handler: router.CreateIdentificationServiceArea}

	pattern = regexp.MustCompile("^/rid_v1/v1/dss/identification_service_areas/(?P<id>[^/]*)/(?P<version>[^/]*)$")
	router.Routes[3] = &api.Route{Pattern: pattern, Handler: router.UpdateIdentificationServiceArea}

	pattern = regexp.MustCompile("^/rid_v1/v1/dss/identification_service_areas/(?P<id>[^/]*)/(?P<version>[^/]*)$")
	router.Routes[4] = &api.Route{Pattern: pattern, Handler: router.DeleteIdentificationServiceArea}

	pattern = regexp.MustCompile("^/rid_v1/v1/dss/subscriptions$")
	router.Routes[5] = &api.Route{Pattern: pattern, Handler: router.SearchSubscriptions}

	pattern = regexp.MustCompile("^/rid_v1/v1/dss/subscriptions/(?P<id>[^/]*)$")
	router.Routes[6] = &api.Route{Pattern: pattern, Handler: router.GetSubscription}

	pattern = regexp.MustCompile("^/rid_v1/v1/dss/subscriptions/(?P<id>[^/]*)$")
	router.Routes[7] = &api.Route{Pattern: pattern, Handler: router.CreateSubscription}

	pattern = regexp.MustCompile("^/rid_v1/v1/dss/subscriptions/(?P<id>[^/]*)/(?P<version>[^/]*)$")
	router.Routes[8] = &api.Route{Pattern: pattern, Handler: router.UpdateSubscription}

	pattern = regexp.MustCompile("^/rid_v1/v1/dss/subscriptions/(?P<id>[^/]*)/(?P<version>[^/]*)$")
	router.Routes[9] = &api.Route{Pattern: pattern, Handler: router.DeleteSubscription}

	return router
}
