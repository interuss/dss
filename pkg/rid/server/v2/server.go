package server

import (
	"github.com/robfig/cron/v3"

	"github.com/interuss/dss/pkg/rid/application"
)

// Server implements ridv2.Implementation.
type Server struct {
	App               application.App
	Locality          string
	AllowHTTPBaseUrls bool
	Cron              *cron.Cron
}
