// This file is auto-generated; do not change as any changes will be overwritten
package auxv1

type VersionResponse struct {
	// The version of the DSS.
	Version string `json:"version"`
}

type ErrorResponse struct {
	// Human-readable message indicating what error occurred and/or why.
	Message *string `json:"message,omitempty"`
}

type PoolResponse struct {
	// Identifier of the DSS Airspace Representation shared by the pool of DSS instances to which this DSS instance belongs. Each DSS instance participating in the pool should indicate the same DAR ID as this ID describes the DAR shared by the pool.
	DarId *string `json:"dar_id,omitempty"`
}

type DSSInstancesResponse struct {
	DssInstances *[]DSSInstance `json:"dss_instances,omitempty"`
}

type DSSInstance struct {
	// Identity of this DSS instance participating in the pool (locality).
	Id string `json:"id"`

	// Most recent heartbeat registered for this DSS instance.
	MostRecentHeartbeat *Heartbeat `json:"most_recent_heartbeat,omitempty"`
}

type Heartbeat struct {
	// Time at which heartbeat was registered.
	Timestamp string `json:"timestamp"`

	// Identity (via access token `sub` claim) of client reporting the heartbeat, or omitted if no client reported the heartbeat.
	Reporter *string `json:"reporter,omitempty"`

	// Source/trigger of this heartbeat.
	Source string `json:"source"`

	// Index of this heartbeat within the set of all heartbeats for this pool participant.
	Index *int64 `json:"index,omitempty"`

	// The time by which a new heartbeat should be registered for this DSS instance if the DSS instance operator's system is behaving correctly.
	NextHeartbeatExpectedBefore *string `json:"next_heartbeat_expected_before,omitempty"`
}

type CAsResponse struct {
	// A list of certificates, each in PEM format.
	Cas *[]string `json:"CAs,omitempty"`
}
