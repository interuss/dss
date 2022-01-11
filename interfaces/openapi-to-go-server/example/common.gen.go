// This file is auto-generated; do not change as any changes will be overwritten
package main

import (
	"net/http"
	"regexp"
)

type Handler func(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request)

type Route struct {
	Pattern *regexp.Regexp
	Handler Handler
}

type Router struct {
	Routes         []*Route
	Implementation Implementation
}

func (s *Router) Handle(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range s.Routes {
		if route.Pattern.MatchString(r.URL.Path) {
			route.Handler(route.Pattern, w, r)
			return true
		}
	}
	return false
}

type MultiRouter struct {
	Routers []*Router
}

func (m *MultiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, router := range m.Routers {
		if router.Handle(w, r) {
			return
		}
	}
	http.NotFound(w, r)
}
