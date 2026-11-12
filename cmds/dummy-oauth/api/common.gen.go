// This file is auto-generated; do not change as any changes will be overwritten
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"regexp"
	"sync"
)

// --- Interface definitions ---

type EmptyResponseBody struct{}

type InternalServerErrorBody struct {
	ErrorMessage string `json:"error_message"`
}

// --- Authorization definitions ---

type SecurityScheme string

type RequiredScope string

// All RequiredScope of all SecurityScheme must be presented simultaneously to use this option.
type AuthorizationOption map[SecurityScheme][]RequiredScope

type AuthorizationResult struct {
	// ID of the client making the operation request
	ClientID *string

	// Scopes granted to client making the operation request
	Scopes []string

	// If authorization was not successful, the problem with the authorization
	Error error
}

type Authorizer interface {
	Authorize(w http.ResponseWriter, r *http.Request, authOptions []AuthorizationOption) AuthorizationResult
}

// --- Utilities ---

func WriteJSON(w http.ResponseWriter, code int, obj any) {
	// A 204 response has no body
	if code == http.StatusNoContent {
		w.WriteHeader(code)
		return
	}
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
	Method  string
	Pattern *regexp.Regexp
	Handler Handler
	Name    string
	Path    string
}

type PartialRouter interface {
	Handle(w http.ResponseWriter, r *http.Request) bool
}

// --- Multi-router definitions ---

type MultiRouter struct {
	routers map[string]PartialRouter
	mu      sync.RWMutex
}

func (m *MultiRouter) SetRouter(name string, r PartialRouter) {
	m.mu.Lock()
	if m.routers == nil {
		m.routers = make(map[string]PartialRouter)
	}
	m.routers[name] = r
	m.mu.Unlock()
}

func (m *MultiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	for router := range maps.Values(m.routers) {
		if router.Handle(w, r) {
			m.mu.RUnlock()
			return
		}
	}
	m.mu.RUnlock()
	http.NotFound(w, r)
}
