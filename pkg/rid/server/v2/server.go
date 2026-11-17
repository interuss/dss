package server

import (
	"github.com/interuss/dss/pkg/rid/store"
)

// Server implements ridv2.Implementation.
type Server struct {
	Store             store.Store
	Locality          string
	AllowHTTPBaseUrls bool
}
