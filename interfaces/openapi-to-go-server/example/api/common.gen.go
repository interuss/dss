// This file is auto-generated; do not change as any changes will be overwritten
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

// --- Interface definitions ---

type EmptyResponseBody struct{}

type InternalServerErrorBody struct {
	ErrorMessage string `json:"error_message"`
}

// --- Authorization definitions ---

type AuthorizationOption struct {
	// All of these scopes must be presented simultaneously to use this option
	RequiredScopes []string
}

type SecurityScheme []AuthorizationOption

type AuthorizationResult struct {
	// ID of the client making the operation request
	ClientID *string

	// Scopes granted to client making the operation request
	Scopes []string

	// If authorization was not successful, the problem with the authorization
	Error error
}

type Authorizer interface {
	Authorize(w http.ResponseWriter, r *http.Request, schemes *map[string]SecurityScheme) AuthorizationResult
}

// --- Utilities ---

func WriteJSON(w http.ResponseWriter, code int, obj interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		if _, err = io.WriteString(w, fmt.Sprintf("{\"error_message\": \"Error encoding JSON: %s\"}", err.Error())); err != nil {
			log.Panicf("Unable to encode JSON for %d response: %v", code, err)
		}
	}
}

// --- API router definitions ---

type Handler func(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request)

type Route struct {
	Pattern *regexp.Regexp
	Handler Handler
}

type PartialRouter interface {
	Handle(w http.ResponseWriter, r *http.Request) bool
}

// --- Multi-router definitions ---

type MultiRouter struct {
	Routers []PartialRouter
}

func (m *MultiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, router := range m.Routers {
		if router.Handle(w, r) {
			return
		}
	}
	http.NotFound(w, r)
}
