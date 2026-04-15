package v1

import (
	"github.com/interuss/dss/pkg/rid/application"
)

// Server implements ridv1.Implementation.
type Server struct {
	App               application.App
	Locality          string
	AllowHTTPBaseUrls bool
}
