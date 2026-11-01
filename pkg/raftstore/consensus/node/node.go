// Package node describes the DSS nodes of a Raft group and tracks the state of this node in its group.
package node

import (
	"sync"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	"go.etcd.io/raft/v3"
)

// State is the state of a DSS node in a Raft group.
type State string

const (
	Waiting State = "waiting" // waiting for the Raft group to be formed or for this node to be added to the group
	Member  State = "member"
	Removed State = "removed"
)

// Info describes a DSS node in a Raft group.
// Its JSON form is what the DSS nodes exchange on the Raft listener.
type Info struct {
	ID    uint64 `json:"id"`
	URL   string `json:"url"`
	State State  `json:"state"`
}

// NodeState holds the raft node along with the state of this node in the group. It is safe for concurrent use.
type NodeState struct {
	mu    sync.RWMutex
	state State
	node  raft.Node
}

// NewNodeState returns the NodeState of a node that is a member of the group.
func NewNodeState(node raft.Node) *NodeState {
	return &NodeState{state: Member, node: node}
}

// get returns the raft node and the state of this node.
func (t *NodeState) get() (raft.Node, State) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.node, t.state
}

// RaftNode returns the raft node.
func (t *NodeState) RaftNode() raft.Node {
	node, _ := t.get()
	return node
}

// Current returns the state of this node.
func (t *NodeState) Current() State {
	_, state := t.get()
	return state
}

// Member returns the raft node, or a dsserr.Unavailable error if this node is not a member of the group.
func (t *NodeState) Member() (raft.Node, error) {
	node, state := t.get()
	if state != Member {
		return nil, stacktrace.NewErrorWithCode(dsserr.Unavailable, "this node is not a member of the Raft group, its state is %s", state)
	}
	return node, nil
}

// SetRemoved records that this node was removed from the group.
func (t *NodeState) SetRemoved() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state = Removed
}
