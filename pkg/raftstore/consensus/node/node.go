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
	// node is nil until this node is a member of the group. Once set, it never changes.
	node raft.Node
}

// NewNodeState returns the NodeState of a node that is waiting to bootstrap or join the group.
func NewNodeState() *NodeState {
	return &NodeState{state: Waiting}
}

// Get returns the raft node, nil until this node is a member of the group, and the state of this node.
func (t *NodeState) Get() (raft.Node, State) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.node, t.state
}

// RaftNode returns the raft node, nil until this node is a member of the group.
func (t *NodeState) RaftNode() raft.Node {
	node, _ := t.Get()
	return node
}

// Current returns the state of this node.
func (t *NodeState) Current() State {
	_, state := t.Get()
	return state
}

// Member returns the raft node, or a dsserr.Unavailable error if this node is not a member of the group.
func (t *NodeState) Member() (raft.Node, error) {
	node, state := t.Get()
	if state != Member {
		return nil, stacktrace.NewErrorWithCode(dsserr.Unavailable, "this node is not a member of the Raft group, its state is %s", state)
	}
	return node, nil
}

// Start makes this node a member of the group with the raft node returned by newNode, if this node is waiting.
// It returns whether the raft node was started, and the state of this node.
func (t *NodeState) Start(newNode func() raft.Node) (bool, State) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state != Waiting {
		return false, t.state
	}
	t.node = newNode()
	t.state = Member
	return true, t.state
}

// SetRemoved records that this node was removed from the group.
func (t *NodeState) SetRemoved() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state = Removed
}
