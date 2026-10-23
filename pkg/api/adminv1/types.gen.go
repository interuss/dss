// This file is auto-generated; do not change as any changes will be overwritten
package adminv1

type ErrorResponse struct {
	// Human-readable message indicating what error occurred and/or why.
	Message *string `json:"message,omitempty"`
}

// Name of a Raft group of a DSS node, one per service. Examples: aux, rid, scd.
type RaftGroupName string

// A DSS node in a Raft group.
type RaftNode struct {
	// ID of the node in the Raft group.
	Id string `json:"id"`

	// Name of the node which is the same across all Raft groups.
	Name string `json:"name"`

	// URL at which the other members of the Raft group send Raft messages to the node. It is served by the Raft communication listener.
	Url string `json:"url"`

	// State of the node in the Raft group:
	//   * `waiting`: the node is not a member of the group yet. It waits to be added to a running group or to create a new one.
	//   * `member`: the node is a member of the group.
	//   * `removed`: the node was removed from the group. Its ID can never be added to the group again.
	State string `json:"state"`
}

// Request to add another DSS node to a Raft group. This request is sent to a member of the group, which does not know the node to add yet: the client passes it as `node`, as returned by `GET /admin/v1/raft/groups/{group_name}/this_node` on the node to add.
type CreateRaftGroupMemberParameters struct {
	Node RaftNode `json:"node"`

	// ID of a node with the same name and in the `member` state that the node replaces, e.g. after it lost its data directory. Without it, adding a node whose name is already used by a member is rejected. The replaced member is not removed by this request: it must be removed once the node is added.
	Replaces *string `json:"replaces,omitempty"`
}

// Request to create a Raft group with this DSS node as its only member. Used for bootstrapping a new Raft group, which is then joined by the other nodes one at a time.
type CreateRaftGroupParameters struct {
	GroupName RaftGroupName `json:"group_name"`
}

// A Raft group, as known by this DSS node.
type RaftGroup struct {
	GroupName RaftGroupName `json:"group_name"`

	// ID of the Raft group within the cluster. It is identical on all the members of the group.
	ClusterId string `json:"cluster_id"`

	// ID of the current leader of the Raft group, empty if unknown.
	LeaderId string `json:"leader_id"`

	// Nodes that are or were members of the Raft group.
	Nodes []RaftNode `json:"nodes"`
}
