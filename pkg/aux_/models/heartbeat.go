package models

import (
	"time"
)

// Heartbeat represents a DSS instance heartbeat.
type Heartbeat struct {
	Locality                    string
	Source                      string
	Timestamp                   *time.Time
	NextHeartbeatExpectedBefore *time.Time
	Reporter                    string
}
