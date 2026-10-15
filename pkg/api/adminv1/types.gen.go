// This file is auto-generated; do not change as any changes will be overwritten
package adminv1

type ErrorResponse struct {
	// Human-readable message indicating what error occurred and/or why.
	Message *string `json:"message,omitempty"`
}

type RaftMember struct {
	// ID of the node in the Raft group, its `--raft_node_id`.
	Id int64 `json:"id"`

	// URL at which the other members of the Raft group reach the node.
	Url string `json:"url"`
}

type RaftGroupMembers struct {
	// ID of this DSS node in the Raft group.
	NodeId int64 `json:"node_id"`

	// ID of the current leader of the Raft group, 0 if unknown.
	LeaderId int64 `json:"leader_id"`

	// Members of the Raft group, all of which are voters.
	Members []RaftMember `json:"members"`

	// IDs of the nodes removed from the Raft group, which can never be added again to it.
	RemovedIds []int64 `json:"removed_ids"`
}
