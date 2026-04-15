// This file is auto-generated; do not change as any changes will be overwritten
package scdv1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/interuss/dss/pkg/api"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
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

var tracer = otel.Tracer("scdv1.api")

// *scdv1.APIRouter (type defined above) implements the api.PartialRouter interface
func (s *APIRouter) Handle(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range s.Routes {
		if route.Method == r.Method && route.Pattern.MatchString(r.URL.Path) {

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

func setAuthError(ctx context.Context, authErr error, resp401 **ErrorResponse, resp403 **ErrorResponse, resp500 **api.InternalServerErrorBody) {
	switch stacktrace.GetCode(authErr) {
	case dsserr.Unauthenticated:
		*resp401 = &ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authentication failed"))}
	case dsserr.PermissionDenied:
		*resp403 = &ErrorResponse{Message: dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Authorization failed"))}
	default:

		if authErr == nil {
			authErr = stacktrace.NewError("Unknown error")
		}

		*resp500 = &api.InternalServerErrorBody{ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(authErr, "Could not perform authorization"))}
	}
}

func (s *APIRouter) QueryOperationalIntentReferences(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req QueryOperationalIntentReferencesRequest
	var response QueryOperationalIntentReferencesResponseSet

	// Parse request body
	req.Body = new(QueryOperationalIntentReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, QueryOperationalIntentReferencesSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.QueryOperationalIntentReferences(ctx, &req)
	}

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
	var response GetOperationalIntentReferenceResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetOperationalIntentReferenceSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.GetOperationalIntentReference(ctx, &req)
	}

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
	var response CreateOperationalIntentReferenceResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])

	// Parse request body
	req.Body = new(PutOperationalIntentReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, CreateOperationalIntentReferenceSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.CreateOperationalIntentReference(ctx, &req)
	}

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
	var response UpdateOperationalIntentReferenceResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])
	req.Ovn = EntityOVN(pathMatch[2])

	// Parse request body
	req.Body = new(PutOperationalIntentReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, UpdateOperationalIntentReferenceSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.UpdateOperationalIntentReference(ctx, &req)
	}

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
	var response DeleteOperationalIntentReferenceResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])
	req.Ovn = EntityOVN(pathMatch[2])

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, DeleteOperationalIntentReferenceSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.DeleteOperationalIntentReference(ctx, &req)
	}

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
	var response QueryConstraintReferencesResponseSet

	// Parse request body
	req.Body = new(QueryConstraintReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, QueryConstraintReferencesSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.QueryConstraintReferences(ctx, &req)
	}

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
	var response GetConstraintReferenceResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetConstraintReferenceSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.GetConstraintReference(ctx, &req)
	}

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
	var response CreateConstraintReferenceResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])

	// Parse request body
	req.Body = new(PutConstraintReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, CreateConstraintReferenceSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.CreateConstraintReference(ctx, &req)
	}

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
	var response UpdateConstraintReferenceResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])
	req.Ovn = EntityOVN(pathMatch[2])

	// Parse request body
	req.Body = new(PutConstraintReferenceParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, UpdateConstraintReferenceSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.UpdateConstraintReference(ctx, &req)
	}

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
	var response DeleteConstraintReferenceResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Entityid = EntityID(pathMatch[1])
	req.Ovn = EntityOVN(pathMatch[2])

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, DeleteConstraintReferenceSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.DeleteConstraintReference(ctx, &req)
	}

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
	var response QuerySubscriptionsResponseSet

	// Parse request body
	req.Body = new(QuerySubscriptionParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, QuerySubscriptionsSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.QuerySubscriptions(ctx, &req)
	}

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
	var response GetSubscriptionResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Subscriptionid = SubscriptionID(pathMatch[1])

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetSubscriptionSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.GetSubscription(ctx, &req)
	}

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
	var response CreateSubscriptionResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Subscriptionid = SubscriptionID(pathMatch[1])

	// Parse request body
	req.Body = new(PutSubscriptionParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, CreateSubscriptionSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.CreateSubscription(ctx, &req)
	}

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
	var response UpdateSubscriptionResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Subscriptionid = SubscriptionID(pathMatch[1])
	req.Version = pathMatch[2]

	// Parse request body
	req.Body = new(PutSubscriptionParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, UpdateSubscriptionSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.UpdateSubscription(ctx, &req)
	}

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
	var response DeleteSubscriptionResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.Subscriptionid = SubscriptionID(pathMatch[1])
	req.Version = pathMatch[2]

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, DeleteSubscriptionSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.DeleteSubscription(ctx, &req)
	}

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
	var response MakeDssReportResponseSet

	// Parse request body
	req.Body = new(ErrorReport)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, MakeDssReportSecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.MakeDssReport(ctx, &req)
	}

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
	var response GetUssAvailabilityResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.UssId = pathMatch[1]

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, GetUssAvailabilitySecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.GetUssAvailability(ctx, &req)
	}

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
	var response SetUssAvailabilityResponseSet

	// Parse path parameters
	pathMatch := exp.FindStringSubmatch(r.URL.Path)
	req.UssId = pathMatch[1]

	// Parse request body
	req.Body = new(SetUssAvailabilityStatusParameters)
	defer r.Body.Close()
	req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, SetUssAvailabilitySecurity)
	// Verify authorization
	if req.Auth.Error != nil {
		setAuthError(r.Context(), stacktrace.Propagate(req.Auth.Error, "Auth failed"), &response.Response401, &response.Response403, &response.Response500)
	} else {
		// Call implementation
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		response = s.Implementation.SetUssAvailability(ctx, &req)
	}

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

func MakeAPIRouter(impl Implementation, auth api.Authorizer) APIRouter {
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 18)}

	pattern := regexp.MustCompile("^/dss/v1/operational_intent_references/query$")
	router.Routes[0] = &api.Route{Method: http.MethodPost, Pattern: pattern, Handler: router.QueryOperationalIntentReferences, Name: "scdv1.QueryOperationalIntentReferences", Path: "/dss/v1/operational_intent_references/query"}

	pattern = regexp.MustCompile("^/dss/v1/operational_intent_references/(?P<entityid>[^/]*)$")
	router.Routes[1] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetOperationalIntentReference, Name: "scdv1.GetOperationalIntentReference", Path: "/dss/v1/operational_intent_references/{entityid}"}

	pattern = regexp.MustCompile("^/dss/v1/operational_intent_references/(?P<entityid>[^/]*)$")
	router.Routes[2] = &api.Route{Method: http.MethodPut, Pattern: pattern, Handler: router.CreateOperationalIntentReference, Name: "scdv1.CreateOperationalIntentReference", Path: "/dss/v1/operational_intent_references/{entityid}"}

	pattern = regexp.MustCompile("^/dss/v1/operational_intent_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[3] = &api.Route{Method: http.MethodPut, Pattern: pattern, Handler: router.UpdateOperationalIntentReference, Name: "scdv1.UpdateOperationalIntentReference", Path: "/dss/v1/operational_intent_references/{entityid}/{ovn}"}

	pattern = regexp.MustCompile("^/dss/v1/operational_intent_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[4] = &api.Route{Method: http.MethodDelete, Pattern: pattern, Handler: router.DeleteOperationalIntentReference, Name: "scdv1.DeleteOperationalIntentReference", Path: "/dss/v1/operational_intent_references/{entityid}/{ovn}"}

	pattern = regexp.MustCompile("^/dss/v1/constraint_references/query$")
	router.Routes[5] = &api.Route{Method: http.MethodPost, Pattern: pattern, Handler: router.QueryConstraintReferences, Name: "scdv1.QueryConstraintReferences", Path: "/dss/v1/constraint_references/query"}

	pattern = regexp.MustCompile("^/dss/v1/constraint_references/(?P<entityid>[^/]*)$")
	router.Routes[6] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetConstraintReference, Name: "scdv1.GetConstraintReference", Path: "/dss/v1/constraint_references/{entityid}"}

	pattern = regexp.MustCompile("^/dss/v1/constraint_references/(?P<entityid>[^/]*)$")
	router.Routes[7] = &api.Route{Method: http.MethodPut, Pattern: pattern, Handler: router.CreateConstraintReference, Name: "scdv1.CreateConstraintReference", Path: "/dss/v1/constraint_references/{entityid}"}

	pattern = regexp.MustCompile("^/dss/v1/constraint_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[8] = &api.Route{Method: http.MethodPut, Pattern: pattern, Handler: router.UpdateConstraintReference, Name: "scdv1.UpdateConstraintReference", Path: "/dss/v1/constraint_references/{entityid}/{ovn}"}

	pattern = regexp.MustCompile("^/dss/v1/constraint_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[9] = &api.Route{Method: http.MethodDelete, Pattern: pattern, Handler: router.DeleteConstraintReference, Name: "scdv1.DeleteConstraintReference", Path: "/dss/v1/constraint_references/{entityid}/{ovn}"}

	pattern = regexp.MustCompile("^/dss/v1/subscriptions/query$")
	router.Routes[10] = &api.Route{Method: http.MethodPost, Pattern: pattern, Handler: router.QuerySubscriptions, Name: "scdv1.QuerySubscriptions", Path: "/dss/v1/subscriptions/query"}

	pattern = regexp.MustCompile("^/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)$")
	router.Routes[11] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetSubscription, Name: "scdv1.GetSubscription", Path: "/dss/v1/subscriptions/{subscriptionid}"}

	pattern = regexp.MustCompile("^/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)$")
	router.Routes[12] = &api.Route{Method: http.MethodPut, Pattern: pattern, Handler: router.CreateSubscription, Name: "scdv1.CreateSubscription", Path: "/dss/v1/subscriptions/{subscriptionid}"}

	pattern = regexp.MustCompile("^/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)/(?P<version>[^/]*)$")
	router.Routes[13] = &api.Route{Method: http.MethodPut, Pattern: pattern, Handler: router.UpdateSubscription, Name: "scdv1.UpdateSubscription", Path: "/dss/v1/subscriptions/{subscriptionid}/{version}"}

	pattern = regexp.MustCompile("^/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)/(?P<version>[^/]*)$")
	router.Routes[14] = &api.Route{Method: http.MethodDelete, Pattern: pattern, Handler: router.DeleteSubscription, Name: "scdv1.DeleteSubscription", Path: "/dss/v1/subscriptions/{subscriptionid}/{version}"}

	pattern = regexp.MustCompile("^/dss/v1/reports$")
	router.Routes[15] = &api.Route{Method: http.MethodPost, Pattern: pattern, Handler: router.MakeDssReport, Name: "scdv1.MakeDssReport", Path: "/dss/v1/reports"}

	pattern = regexp.MustCompile("^/dss/v1/uss_availability/(?P<uss_id>[^/]*)$")
	router.Routes[16] = &api.Route{Method: http.MethodGet, Pattern: pattern, Handler: router.GetUssAvailability, Name: "scdv1.GetUssAvailability", Path: "/dss/v1/uss_availability/{uss_id}"}

	pattern = regexp.MustCompile("^/dss/v1/uss_availability/(?P<uss_id>[^/]*)$")
	router.Routes[17] = &api.Route{Method: http.MethodPut, Pattern: pattern, Handler: router.SetUssAvailability, Name: "scdv1.SetUssAvailability", Path: "/dss/v1/uss_availability/{uss_id}"}

	return router
}
