// This file is auto-generated; do not change as any changes will be overwritten
package scd_v1

import (
	"context"
	"encoding/json"
	"github.com/interuss/dss/pkg/api"
	"net/http"
	"regexp"
)

type APIRouter struct {
	Routes         []*api.Route
	Implementation Implementation
	Authorizer     api.Authorizer
}

// *scd_v1.APIRouter (type defined above) implements the api.PartialRouter interface
func (s *APIRouter) Handle(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range s.Routes {
		if route.Pattern.MatchString(r.URL.Path) {
			route.Handler(route.Pattern, w, r)
			return true
		}
	}
	return false
}

func (s *APIRouter) QueryOperationalIntentReferences(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req QueryOperationalIntentReferencesRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &QueryOperationalIntentReferencesSecurity)

	// Parse request body
	req.Body = new(QueryOperationalIntentReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.QueryOperationalIntentReferences(ctx, &req)

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

func (s *APIRouter) GetOperationalIntentReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetOperationalIntentReferenceRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &GetOperationalIntentReferenceSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.GetOperationalIntentReference(ctx, &req)

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

func (s *APIRouter) CreateOperationalIntentReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req CreateOperationalIntentReferenceRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &CreateOperationalIntentReferenceSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])

	// Parse request body
	req.Body = new(PutOperationalIntentReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.CreateOperationalIntentReference(ctx, &req)

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
	if response.Response409 != nil {
		api.WriteJSON(w, 409, response.Response409)
		return
	}
	if response.Response412 != nil {
		api.WriteJSON(w, 412, response.Response412)
		return
	}
	if response.Response413 != nil {
		api.WriteJSON(w, 413, response.Response413)
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

func (s *APIRouter) UpdateOperationalIntentReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req UpdateOperationalIntentReferenceRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &UpdateOperationalIntentReferenceSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])
	req.Ovn = EntityOVN(pathMatch[2])

	// Parse request body
	req.Body = new(PutOperationalIntentReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.UpdateOperationalIntentReference(ctx, &req)

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
	if response.Response412 != nil {
		api.WriteJSON(w, 412, response.Response412)
		return
	}
	if response.Response413 != nil {
		api.WriteJSON(w, 413, response.Response413)
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

func (s *APIRouter) DeleteOperationalIntentReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req DeleteOperationalIntentReferenceRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &DeleteOperationalIntentReferenceSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])
	req.Ovn = EntityOVN(pathMatch[2])

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.DeleteOperationalIntentReference(ctx, &req)

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
	if response.Response412 != nil {
		api.WriteJSON(w, 412, response.Response412)
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

func (s *APIRouter) QueryConstraintReferences(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req QueryConstraintReferencesRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &QueryConstraintReferencesSecurity)

	// Parse request body
	req.Body = new(QueryConstraintReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.QueryConstraintReferences(ctx, &req)

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

func (s *APIRouter) GetConstraintReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetConstraintReferenceRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &GetConstraintReferenceSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.GetConstraintReference(ctx, &req)

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

func (s *APIRouter) CreateConstraintReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req CreateConstraintReferenceRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &CreateConstraintReferenceSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])

	// Parse request body
	req.Body = new(PutConstraintReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.CreateConstraintReference(ctx, &req)

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
	if response.Response409 != nil {
		api.WriteJSON(w, 409, response.Response409)
		return
	}
	if response.Response413 != nil {
		api.WriteJSON(w, 413, response.Response413)
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

func (s *APIRouter) UpdateConstraintReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req UpdateConstraintReferenceRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &UpdateConstraintReferenceSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])
	req.Ovn = EntityOVN(pathMatch[2])

	// Parse request body
	req.Body = new(PutConstraintReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.UpdateConstraintReference(ctx, &req)

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

func (s *APIRouter) DeleteConstraintReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req DeleteConstraintReferenceRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &DeleteConstraintReferenceSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])
	req.Ovn = EntityOVN(pathMatch[2])

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.DeleteConstraintReference(ctx, &req)

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

func (s *APIRouter) QuerySubscriptions(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req QuerySubscriptionsRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &QuerySubscriptionsSecurity)

	// Parse request body
	req.Body = new(QuerySubscriptionParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.QuerySubscriptions(ctx, &req)

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

func (s *APIRouter) GetSubscription(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetSubscriptionRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &GetSubscriptionSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Subscriptionid = SubscriptionID(pathMatch[1])

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

func (s *APIRouter) CreateSubscription(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &CreateSubscriptionSecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Subscriptionid = SubscriptionID(pathMatch[1])

	// Parse request body
	req.Body = new(PutSubscriptionParameters)
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
	req.Subscriptionid = SubscriptionID(pathMatch[1])
	req.Version = pathMatch[2]

	// Parse request body
	req.Body = new(PutSubscriptionParameters)
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
	req.Subscriptionid = SubscriptionID(pathMatch[1])
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

func (s *APIRouter) MakeDssReport(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req MakeDssReportRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &MakeDssReportSecurity)

	// Parse request body
	req.Body = new(ErrorReport)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.MakeDssReport(ctx, &req)

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

func (s *APIRouter) GetUssAvailability(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetUssAvailabilityRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &GetUssAvailabilitySecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.UssId = pathMatch[1]

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.GetUssAvailability(ctx, &req)

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

func (s *APIRouter) SetUssAvailability(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req SetUssAvailabilityRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &SetUssAvailabilitySecurity)

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.UssId = pathMatch[1]

	// Parse request body
	req.Body = new(SetUssAvailabilityStatusParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.SetUssAvailability(ctx, &req)

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

func MakeAPIRouter(impl Implementation, auth api.Authorizer) APIRouter {
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 18)}

	pattern := regexp.MustCompile("^/scd_v1/dss/v1/operational_intent_references/query$")
	router.Routes[0] = &api.Route{Pattern: pattern, Handler: router.QueryOperationalIntentReferences}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/operational_intent_references/(?P<entityid>[^/]*)$")
	router.Routes[1] = &api.Route{Pattern: pattern, Handler: router.GetOperationalIntentReference}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/operational_intent_references/(?P<entityid>[^/]*)$")
	router.Routes[2] = &api.Route{Pattern: pattern, Handler: router.CreateOperationalIntentReference}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/operational_intent_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[3] = &api.Route{Pattern: pattern, Handler: router.UpdateOperationalIntentReference}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/operational_intent_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[4] = &api.Route{Pattern: pattern, Handler: router.DeleteOperationalIntentReference}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/constraint_references/query$")
	router.Routes[5] = &api.Route{Pattern: pattern, Handler: router.QueryConstraintReferences}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/constraint_references/(?P<entityid>[^/]*)$")
	router.Routes[6] = &api.Route{Pattern: pattern, Handler: router.GetConstraintReference}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/constraint_references/(?P<entityid>[^/]*)$")
	router.Routes[7] = &api.Route{Pattern: pattern, Handler: router.CreateConstraintReference}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/constraint_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[8] = &api.Route{Pattern: pattern, Handler: router.UpdateConstraintReference}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/constraint_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[9] = &api.Route{Pattern: pattern, Handler: router.DeleteConstraintReference}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/subscriptions/query$")
	router.Routes[10] = &api.Route{Pattern: pattern, Handler: router.QuerySubscriptions}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)$")
	router.Routes[11] = &api.Route{Pattern: pattern, Handler: router.GetSubscription}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)$")
	router.Routes[12] = &api.Route{Pattern: pattern, Handler: router.CreateSubscription}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)/(?P<version>[^/]*)$")
	router.Routes[13] = &api.Route{Pattern: pattern, Handler: router.UpdateSubscription}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)/(?P<version>[^/]*)$")
	router.Routes[14] = &api.Route{Pattern: pattern, Handler: router.DeleteSubscription}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/reports$")
	router.Routes[15] = &api.Route{Pattern: pattern, Handler: router.MakeDssReport}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/uss_availability/(?P<uss_id>[^/]*)$")
	router.Routes[16] = &api.Route{Pattern: pattern, Handler: router.GetUssAvailability}

	pattern = regexp.MustCompile("^/scd_v1/dss/v1/uss_availability/(?P<uss_id>[^/]*)$")
	router.Routes[17] = &api.Route{Pattern: pattern, Handler: router.SetUssAvailability}

	return router
}
