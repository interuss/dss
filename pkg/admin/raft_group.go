package admin

import "context"

// RaftGroup manages the membership of a Raft group.
type RaftGroup interface {
	// Members returns the membership of the group as known by the DSS node.
	Members(ctx context.Context) (Membership, error)
	// AddNode adds nodeID, reachable by the other members at peerURL, to the group.
	// It returns the membership of the group after the change is applied.
	AddNode(ctx context.Context, nodeID uint64, peerURL string) (Membership, error)
	// RemoveNode removes nodeID from the group. Its ID is retired and can never be reused.
	// It returns the membership of the group after the change is applied.
	RemoveNode(ctx context.Context, nodeID uint64) (Membership, error)
}

// Membership describes the membership of a Raft group, as known by a node.
type Membership struct {
	// NodeID is the ID of the node reporting the membership.
	NodeID uint64
	// LeaderID is the ID of the current leader of the group, 0 if unknown.
	LeaderID uint64
	// Members maps the ID of every member of the group to its peer URL.
	Members map[uint64]string
	// RemovedIDs are the IDs of the nodes removed from the group.
	RemovedIDs []uint64
}
