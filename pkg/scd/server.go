package scd

import (
	scdstore "github.com/interuss/dss/pkg/scd/store"
)

// Server implements scdv1.Implementation.
type Server struct {
	Store             scdstore.Store
	DSSReportHandler  ReceivedReportHandler
	AllowHTTPBaseUrls bool
}
