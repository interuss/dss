// This file is auto-generated; do not change as any changes will be overwritten
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

func writeJson(w http.ResponseWriter, code int, obj interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		io.WriteString(w, fmt.Sprintf("{\"error_message\": \"Error encoding JSON: %s\"}", err.Error()))
	}
}

func (s *Router) QueryOperationalIntentReferences(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var bodyParseError *error
	body := new(QueryOperationalIntentReferenceParameters)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.QueryOperationalIntentReferences(*body, bodyParseError)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response413 != nil {
		writeJson(w, 413, response.Response413)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) CreateOperationalIntentReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	entityid := EntityID(path_match[1])

	// Parse request body
	var bodyParseError *error
	body := new(PutOperationalIntentReferenceParameters)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.CreateOperationalIntentReference(entityid, *body, bodyParseError)

	// Write response to client
	if response.Response201 != nil {
		writeJson(w, 201, response.Response201)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response409 != nil {
		writeJson(w, 409, response.Response409)
		return
	}
	if response.Response412 != nil {
		writeJson(w, 412, response.Response412)
		return
	}
	if response.Response413 != nil {
		writeJson(w, 413, response.Response413)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) GetOperationalIntentReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	entityid := EntityID(path_match[1])

	// Call implementation
	response := s.Implementation.GetOperationalIntentReference(entityid)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response404 != nil {
		writeJson(w, 404, response.Response404)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) DeleteOperationalIntentReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	entityid := EntityID(path_match[1])
	ovn := EntityOVN(path_match[2])

	// Call implementation
	response := s.Implementation.DeleteOperationalIntentReference(entityid, ovn)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response404 != nil {
		writeJson(w, 404, response.Response404)
		return
	}
	if response.Response409 != nil {
		writeJson(w, 409, response.Response409)
		return
	}
	if response.Response412 != nil {
		writeJson(w, 412, response.Response412)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) UpdateOperationalIntentReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	entityid := EntityID(path_match[1])
	ovn := EntityOVN(path_match[2])

	// Parse request body
	var bodyParseError *error
	body := new(PutOperationalIntentReferenceParameters)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.UpdateOperationalIntentReference(entityid, ovn, *body, bodyParseError)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response409 != nil {
		writeJson(w, 409, response.Response409)
		return
	}
	if response.Response412 != nil {
		writeJson(w, 412, response.Response412)
		return
	}
	if response.Response413 != nil {
		writeJson(w, 413, response.Response413)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) QueryConstraintReferences(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var bodyParseError *error
	body := new(QueryConstraintReferenceParameters)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.QueryConstraintReferences(*body, bodyParseError)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response413 != nil {
		writeJson(w, 413, response.Response413)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) CreateConstraintReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	entityid := EntityID(path_match[1])

	// Parse request body
	var bodyParseError *error
	body := new(PutConstraintReferenceParameters)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.CreateConstraintReference(entityid, *body, bodyParseError)

	// Write response to client
	if response.Response201 != nil {
		writeJson(w, 201, response.Response201)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response409 != nil {
		writeJson(w, 409, response.Response409)
		return
	}
	if response.Response413 != nil {
		writeJson(w, 413, response.Response413)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) GetConstraintReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	entityid := EntityID(path_match[1])

	// Call implementation
	response := s.Implementation.GetConstraintReference(entityid)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response404 != nil {
		writeJson(w, 404, response.Response404)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) DeleteConstraintReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	entityid := EntityID(path_match[1])
	ovn := EntityOVN(path_match[2])

	// Call implementation
	response := s.Implementation.DeleteConstraintReference(entityid, ovn)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response404 != nil {
		writeJson(w, 404, response.Response404)
		return
	}
	if response.Response409 != nil {
		writeJson(w, 409, response.Response409)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) UpdateConstraintReference(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	entityid := EntityID(path_match[1])
	ovn := EntityOVN(path_match[2])

	// Parse request body
	var bodyParseError *error
	body := new(PutConstraintReferenceParameters)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.UpdateConstraintReference(entityid, ovn, *body, bodyParseError)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response409 != nil {
		writeJson(w, 409, response.Response409)
		return
	}
	if response.Response413 != nil {
		writeJson(w, 413, response.Response413)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) QuerySubscriptions(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var bodyParseError *error
	body := new(QuerySubscriptionParameters)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.QuerySubscriptions(*body, bodyParseError)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response413 != nil {
		writeJson(w, 413, response.Response413)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) CreateSubscription(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	subscriptionid := SubscriptionID(path_match[1])

	// Parse request body
	var bodyParseError *error
	body := new(PutSubscriptionParameters)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.CreateSubscription(subscriptionid, *body, bodyParseError)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response409 != nil {
		writeJson(w, 409, response.Response409)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) GetSubscription(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	subscriptionid := SubscriptionID(path_match[1])

	// Call implementation
	response := s.Implementation.GetSubscription(subscriptionid)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response404 != nil {
		writeJson(w, 404, response.Response404)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) DeleteSubscription(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	subscriptionid := SubscriptionID(path_match[1])
	version := path_match[2]

	// Call implementation
	response := s.Implementation.DeleteSubscription(subscriptionid, version)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response404 != nil {
		writeJson(w, 404, response.Response404)
		return
	}
	if response.Response409 != nil {
		writeJson(w, 409, response.Response409)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) UpdateSubscription(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	subscriptionid := SubscriptionID(path_match[1])
	version := path_match[2]

	// Parse request body
	var bodyParseError *error
	body := new(PutSubscriptionParameters)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.UpdateSubscription(subscriptionid, version, *body, bodyParseError)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response409 != nil {
		writeJson(w, 409, response.Response409)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) MakeDssReport(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var bodyParseError *error
	body := new(ErrorReport)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.MakeDssReport(*body, bodyParseError)

	// Write response to client
	if response.Response201 != nil {
		writeJson(w, 201, response.Response201)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) SetUssAvailability(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	uss_id := path_match[1]

	// Parse request body
	var bodyParseError *error
	body := new(SetUssAvailabilityStatusParameters)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		bodyParseError = &err
	}

	// Call implementation
	response := s.Implementation.SetUssAvailability(uss_id, *body, bodyParseError)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func (s *Router) GetUssAvailability(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	// Parse path parameters
	path_match := exp.FindStringSubmatch(r.URL.Path)
	uss_id := path_match[1]

	// Call implementation
	response := s.Implementation.GetUssAvailability(uss_id)

	// Write response to client
	if response.Response200 != nil {
		writeJson(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		writeJson(w, 400, response.Response400)
		return
	}
	if response.Response401 != nil {
		writeJson(w, 401, response.Response401)
		return
	}
	if response.Response403 != nil {
		writeJson(w, 403, response.Response403)
		return
	}
	if response.Response429 != nil {
		writeJson(w, 429, response.Response429)
		return
	}
	if response.Response500 != nil {
		writeJson(w, 500, response.Response500)
		return
	}
	writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})
}

func MakeRouter(impl Implementation) Router {
	router := Router{Implementation: impl, Routes: make([]*Route, 18)}

	pattern := regexp.MustCompile("^/scd/dss/v1/operational_intent_references/query$")
	router.Routes[0] = &Route{Pattern: pattern, Handler: router.QueryOperationalIntentReferences}

	pattern = regexp.MustCompile("^/scd/dss/v1/operational_intent_references/(?P<entityid>[^/]*)$")
	router.Routes[1] = &Route{Pattern: pattern, Handler: router.CreateOperationalIntentReference}

	pattern = regexp.MustCompile("^/scd/dss/v1/operational_intent_references/(?P<entityid>[^/]*)$")
	router.Routes[2] = &Route{Pattern: pattern, Handler: router.GetOperationalIntentReference}

	pattern = regexp.MustCompile("^/scd/dss/v1/operational_intent_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[3] = &Route{Pattern: pattern, Handler: router.DeleteOperationalIntentReference}

	pattern = regexp.MustCompile("^/scd/dss/v1/operational_intent_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[4] = &Route{Pattern: pattern, Handler: router.UpdateOperationalIntentReference}

	pattern = regexp.MustCompile("^/scd/dss/v1/constraint_references/query$")
	router.Routes[5] = &Route{Pattern: pattern, Handler: router.QueryConstraintReferences}

	pattern = regexp.MustCompile("^/scd/dss/v1/constraint_references/(?P<entityid>[^/]*)$")
	router.Routes[6] = &Route{Pattern: pattern, Handler: router.CreateConstraintReference}

	pattern = regexp.MustCompile("^/scd/dss/v1/constraint_references/(?P<entityid>[^/]*)$")
	router.Routes[7] = &Route{Pattern: pattern, Handler: router.GetConstraintReference}

	pattern = regexp.MustCompile("^/scd/dss/v1/constraint_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[8] = &Route{Pattern: pattern, Handler: router.DeleteConstraintReference}

	pattern = regexp.MustCompile("^/scd/dss/v1/constraint_references/(?P<entityid>[^/]*)/(?P<ovn>[^/]*)$")
	router.Routes[9] = &Route{Pattern: pattern, Handler: router.UpdateConstraintReference}

	pattern = regexp.MustCompile("^/scd/dss/v1/subscriptions/query$")
	router.Routes[10] = &Route{Pattern: pattern, Handler: router.QuerySubscriptions}

	pattern = regexp.MustCompile("^/scd/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)$")
	router.Routes[11] = &Route{Pattern: pattern, Handler: router.CreateSubscription}

	pattern = regexp.MustCompile("^/scd/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)$")
	router.Routes[12] = &Route{Pattern: pattern, Handler: router.GetSubscription}

	pattern = regexp.MustCompile("^/scd/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)/(?P<version>[^/]*)$")
	router.Routes[13] = &Route{Pattern: pattern, Handler: router.DeleteSubscription}

	pattern = regexp.MustCompile("^/scd/dss/v1/subscriptions/(?P<subscriptionid>[^/]*)/(?P<version>[^/]*)$")
	router.Routes[14] = &Route{Pattern: pattern, Handler: router.UpdateSubscription}

	pattern = regexp.MustCompile("^/scd/dss/v1/reports$")
	router.Routes[15] = &Route{Pattern: pattern, Handler: router.MakeDssReport}

	pattern = regexp.MustCompile("^/scd/dss/v1/uss_availability/(?P<uss_id>[^/]*)$")
	router.Routes[16] = &Route{Pattern: pattern, Handler: router.SetUssAvailability}

	pattern = regexp.MustCompile("^/scd/dss/v1/uss_availability/(?P<uss_id>[^/]*)$")
	router.Routes[17] = &Route{Pattern: pattern, Handler: router.GetUssAvailability}

	return router
}
