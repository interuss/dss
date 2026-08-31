package v1

import (
	"github.com/interuss/dss/pkg/rid/application"
	"github.com/interuss/dss/pkg/rid/store"
)

// Server implements ridv1.Implementation.
type Server struct {
	Store             store.Store
	App               application.App
	Locality          string
	AllowHTTPBaseUrls bool
}
