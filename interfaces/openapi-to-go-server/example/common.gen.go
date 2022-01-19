// This file is auto-generated; do not change as any changes will be overwritten
package main

import (
	"net/http"
	"regexp"
)

// --- Interface definitions ---

type EmptyResponseBody struct{}

type InternalServerErrorBody struct {
	ErrorMessage string `json:"error_message"`
}

// --- Authorization definitions ---

type AuthorizationResult struct {
	ClientID *string
	Error    error
}

type Authorizer interface {
	Authorize(w http.ResponseWriter, r *http.Request, schemes *map[string]SecurityScheme) AuthorizationResult
}

type AuthorizationOption struct {
	RequiredScopes []string
}

type SecurityScheme []AuthorizationOption

// --- API router definitions ---

type Handler func(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request)

type Route struct {
	Pattern *regexp.Regexp
	Handler Handler
}

type APIRouter struct {
	Routes         []*Route
	Implementation Implementation
	Authorizer     Authorizer
}

func (a *APIRouter) Handle(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range a.Routes {
		if route.Pattern.MatchString(r.URL.Path) {
			route.Handler(route.Pattern, w, r)
			return true
		}
	}
	return false
}

// --- Multi-router definitions ---

type MultiRouter struct {
	Routers []*APIRouter
}

func (m *MultiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, router := range m.Routers {
		if router.Handle(w, r) {
			return
		}
	}
	http.NotFound(w, r)
}
