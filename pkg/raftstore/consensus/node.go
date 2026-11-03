package consensus

// NodeState is the state of a DSS node in a Raft group.
type NodeState string

const (
	NodeWaiting NodeState = "waiting" // waiting for the Raft group to be formed or for this node to be added to the group
	NodeMember  NodeState = "member"
	NodeRemoved NodeState = "removed"
)

// Node describes a DSS node in a Raft group.
// Its JSON form is what the DSS nodes exchange on the Raft listener.
type Node struct {
	ID    uint64    `json:"id"`
	URL   string    `json:"url"`
	State NodeState `json:"state"`
}

// Membership describes a Raft group.
type Membership struct {
	// ClusterID identifies the group within the cluster. It is identical on all the members of the group.
	ClusterID uint64
	// LeaderID is the ID in Nodes of the leader of the group, 0 if it is unknown.
	LeaderID uint64
	// Nodes maps the ID of every node that is or was a member of the group to the node.
	// A node's state tells whether it is a member or was removed.
	// Removed IDs can never be reused.
	Nodes map[uint64]Node
}
