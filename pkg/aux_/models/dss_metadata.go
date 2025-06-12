package models

import (
	"time"
)

// DSSMetadata represents a DSS instance metadata.
type DSSMetadata struct {
	Locality       string
	PublicEndpoint string
	UpdatedAt      *time.Time
}
