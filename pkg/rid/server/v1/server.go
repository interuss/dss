package v1

import (
	"github.com/interuss/dss/pkg/rid/store"
)

// Server implements ridv1.Implementation.
type Server struct {
	Store             store.Store
	Locality          string
	AllowHTTPBaseUrls bool
}
