package admin

import "context"

// RaftGroup manages the membership of a Raft group.
type RaftGroup interface {
	// Node returns the description of this DSS node within the group.
	Node(ctx context.Context) (Node, error)
	// Bootstrap starts a new group with this DSS node as its only member. Only a waiting node can bootstrap a group.
	// It returns the membership of the group once the DSS node is its leader.
	// TODO: Add a list of initial members to bootstrap with, so that the group can be bootstrapped with more than one member.
	Bootstrap(ctx context.Context) (Membership, error)
	// Members returns the membership of the group as known by the DSS node.
	// It returns a dsserr.NotFound error if the DSS node is waiting to bootstrap or join the group.
	Members(ctx context.Context) (Membership, error)
	// AddNode adds the DSS node reachable by the other members at peerURL to the group. The ID of the node is read from
	// the node itself, which must be running and waiting.
	// It returns the added node, as a member of the group, once the change is applied.
	// It returns a dsserr.AlreadyExists error if the node is not waiting or does not run this group in this cluster, or
	// if peerURL is already used by another member, and a dsserr.Unavailable error if the node cannot be reached or the
	// DSS node is not a member of the group.
	AddNode(ctx context.Context, peerURL string) (Node, error)
	// RemoveNode removes nodeID from the group. Its ID is retired and can never be reused.
	// It returns once the change is applied.
	RemoveNode(ctx context.Context, nodeID uint64) error
}

type NodeState string

const (
	NodeWaiting NodeState = "waiting"
	NodeMember  NodeState = "member"
	NodeRemoved NodeState = "removed"
)

// Node describes a DSS node in a Raft group.
type Node struct {
	ID    uint64
	URL   string
	State NodeState
}

type Membership struct {
	// ClusterID identifies the group within the cluster. It is identical on all the members of the group.
	ClusterID uint64
	LeaderID  uint64
	// Nodes maps the ID of every node that is or was a member of the group to the node.
	// A node's state tells whether it is a member or was removed.
	// Removed IDs can never be reused.
	Nodes map[uint64]Node
}
