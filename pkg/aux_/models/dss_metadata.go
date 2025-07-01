package models

import (
	"database/sql"
	"time"
)

// DSSMetadata represents a DSS instance metadata.
type DSSMetadata struct {
	Locality        string
	PublicEndpoint  string
	UpdatedAt       *time.Time
	LatestTimestamp struct {
		Source                      sql.NullString
		Timestamp                   *time.Time
		NextHeartbeatExpectedBefore *time.Time
		Reporter                    sql.NullString
	}
}
