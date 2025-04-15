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

type DSSInstancesResponse struct {
	DssInstances *[]DSSInstance `json:"dss_instances,omitempty"`
}

type DSSInstance struct {
	// Identity of this DSS instance participating in the pool (locality).
	InstanceId string `json:"instance_id"`

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

	// If specified, a sufficiently unique value to verify synchronization between DSS instances with additional certainty.  A UUIDv4 is recommended.
	UniqueValue *string `json:"unique_value,omitempty"`
}
