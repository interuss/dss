// This file is auto-generated; do not change as any changes will be overwritten
package dummyoauth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/interuss/dss/cmds/dummy-oauth/api"
)

type APIRouter struct {
	Routes         []*api.Route
	Implementation Implementation
	Authorizer     api.Authorizer
}

// *dummyoauth.APIRouter (type defined above) implements the api.PartialRouter interface
func (s *APIRouter) Handle(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range s.Routes {
		if (route.Method == r.Method) && route.Pattern.MatchString(r.URL.Path) {
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
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		api.WriteJSON(w, 400, response.Response400)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) PostToken(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req PostTokenRequest = *new(PostTokenRequest)

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &PostTokenSecurity)

	msig := r.Header.Get("x-utm-message-signature")
	req.XUtmMessageSignature = &msig

	msigip := r.Header.Get("x-utm-message-signature-input")
	if msigip != "" {
		req.XUtmMessageSignatureInput = &msigip
	}

	var jwsh xUtmMessageSignatureJoseHeader
	err := json.Unmarshal([]byte(r.Header.Get("x-utm-jws-header")), &jwsh)
	if err != nil {
		log.Printf("Unable to unmarshal x-utm-jws-header %s", err)
	}
	req.XUtmJwsHeader = &jwsh

	// Parse request body
	var body TokenRequestForm
	// defer r.Body.Close()
	// req.BodyParseError = json.NewDecoder(r.Body).Decode(&body)

	cd := r.Header.Get("content_digest")
	if cd != "" {
		req.ContentDigest = &cd
	}

	er := r.ParseForm()
	if er != nil {
		e := "Invalid request `body`"
		eDisc := "Could not parse the body form"
		data := map[string]interface{}{
			"Error":            &e,
			"ErrorDescription": &eDisc,
		}
		api.WriteJSON(w, 400, data)
		return
	}
	body.GrantType = r.FormValue("grant_type")
	body.ClientID = r.FormValue("client_id")
	body.Audience = r.FormValue("audience")
	body.Scope = r.FormValue("scope")
	req.Body = &body
	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.PostToken(ctx, &req)

	// Write response to client
	if response.Response200 != nil {
		api.WriteJSON(w, 200, response.Response200)
		return
	}
	if response.Response400 != nil {
		api.WriteJSON(w, 400, response.Response400)
		return
	}
	if response.Response500 != nil {
		api.WriteJSON(w, 500, response.Response500)
		return
	}
	api.WriteJSON(w, 500, api.InternalServerErrorBody{ErrorMessage: "Handler implementation did not set a response"})
}

func (s *APIRouter) GetWellKnownOauthAuthorizationServer(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetWellKnownOauthAuthorizationServerRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &GetWellKnownOauthAuthorizationServerSecurity)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.GetWellKnownOauthAuthorizationServer(ctx, &req)

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

func (s *APIRouter) GetWellKnownJwksJSON(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {
	var req GetWellKnownJwksJSONRequest

	// Authorize request
	req.Auth = s.Authorizer.Authorize(w, r, &GetWellKnownJwksJSONSecurity)

	// Call implementation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	response := s.Implementation.GetWellKnownJwksJSON(ctx, &req)

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
	router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*api.Route, 4)}

	pattern := regexp.MustCompile("^/token")
	router.Routes[0] = &api.Route{Method: "GET", Pattern: pattern, Handler: router.GetToken}

	pattern = regexp.MustCompile("^/token")
	router.Routes[1] = &api.Route{Method: "POST", Pattern: pattern, Handler: router.PostToken}

	pattern = regexp.MustCompile("^/.well-known/oauth-authorization-server$")
	router.Routes[2] = &api.Route{Method: "GET", Pattern: pattern, Handler: router.GetWellKnownOauthAuthorizationServer}

	pattern = regexp.MustCompile("^/.well-known/jwks.json$")
	router.Routes[3] = &api.Route{Method: "GET", Pattern: pattern, Handler: router.GetWellKnownJwksJSON}

	return router
}
