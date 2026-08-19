package server

import (
	"github.com/robfig/cron/v3"

	"github.com/interuss/dss/pkg/rid/application"
	"github.com/interuss/dss/pkg/rid/store"
)

// Server implements ridv2.Implementation.
type Server struct {
	Store             store.Store
	App               application.App
	Locality          string
	AllowHTTPBaseUrls bool
	Cron              *cron.Cron
}
