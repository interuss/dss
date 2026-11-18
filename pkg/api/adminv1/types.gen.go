// This file is auto-generated; do not change as any changes will be overwritten
package adminv1

type ErrorResponse struct {
	// Human-readable message indicating what error occurred and/or why.
	Message *string `json:"message,omitempty"`
}

// Name of a Raft group of a DSS node, one per service (e.g., `aux`, `rid`, `scd`).
type RaftGroupName string

// A DSS node in a Raft group.
type RaftNode struct {
	// Identifier for the node within the Raft group.
	Id string `json:"id"`

	// URL at which the other members of the Raft group send Raft messages to the node. Served by the Raft communication listener. It identifies the DSS node within the cluster and across all Raft groups.
	Url string `json:"url"`

	// State of the node in the Raft group.
	//
	// * `waiting`: Node is not yet a member of the group. It waits to be added to a running group or to create a new one.
	//
	// * `member`: Node is a member of the group.
	//
	// * `removed`: Node was removed from the group. Its ID can never be added to the group again.
	State string `json:"state"`
}

// Request to add another DSS node to a Raft group. This request is sent to a member of the group, which reads the ID of the node to add from the node itself, at its URL.
type CreateRaftGroupMemberParameters struct {
	// URL at which the members of the Raft group send Raft messages to the node to add.
	Url string `json:"url"`
}

// Request to create a Raft group with this DSS node as its only member. Used for bootstrapping a new Raft group, which is then joined by the other nodes one at a time.
type CreateRaftGroupParameters struct {
	GroupName RaftGroupName `json:"group_name"`
}

// A Raft group, as known by this DSS node.
type RaftGroup struct {
	GroupName RaftGroupName `json:"group_name"`

	// Identifier for the Raft group within the cluster. Identical on all the members of the group.
	ClusterId string `json:"cluster_id"`

	// Identifier for the current leader of the Raft group. Empty if unknown.
	LeaderId string `json:"leader_id"`

	// Nodes that are or were members of the Raft group.
	Nodes []RaftNode `json:"nodes"`
}

// Eviction parameters of a Raft group, applied to each of its entity types.
type EvictionParameters struct {
	// Duration, as a Go duration string, after their end at which entities are expired. Must be positive.
	Ttl string `json:"ttl"`

	// Maximum number of entities evicted per entity type by a cleanup job, 0 for unlimited.
	Limit *int64 `json:"limit,omitempty"`
}

// Eviction settings of a Raft group, used by the cleanup jobs scheduled by its leader.
type EvictionConfig struct {
	// Duration, as a Go duration string, after their end at which entities are expired. Must be positive.
	Interval string `json:"interval"`

	Parameters EvictionParameters `json:"parameters"`
}

// Eviction settings of a Raft group and the last run of its cleanup jobs.
type EvictionState struct {
	Config EvictionConfig `json:"config"`

	// Time of the last successful cleanup job of the group, absent if none ran.
	LastRun *string `json:"last_run,omitempty"`
}

// Parameters of a cleanup job run on demand.
type EvictionRunParameters struct {
	// Eviction parameters of the job. When absent, the parameters of the eviction settings of the group are used.
	Parameters *EvictionParameters `json:"parameters,omitempty"`
}

// Number of entities of an entity type deleted by a cleanup job.
type EntityEvictionCount struct {
	// Entity type, e.g. `isas`.
	Type string `json:"type"`

	// Number of entities deleted.
	Count int64 `json:"count"`
}

// Result of a cleanup job run on demand.
type EvictionRunResult struct {
	// Number of entities deleted per entity type of the group, sorted by entity type.
	Counts []EntityEvictionCount `json:"counts"`
}
