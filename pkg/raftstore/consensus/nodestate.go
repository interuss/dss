package consensus

import (
	"sync"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	"go.etcd.io/raft/v3"
)

type nodeState struct {
	sync.RWMutex
	state NodeState
	// node is nil until this node is a member of the group. Once set, it never changes.
	node raft.Node
}

// get returns the raft node, nil until this node is a member of the group, and the state of this node.
func (s *nodeState) get() (raft.Node, NodeState) {
	s.RLock()
	defer s.RUnlock()
	return s.node, s.state
}

// raftNode returns the raft node, nil until this node is a member of the group.
func (s *nodeState) raftNode() raft.Node {
	node, _ := s.get()
	return node
}

// current returns the state of this node.
func (s *nodeState) current() NodeState {
	_, state := s.get()
	return state
}

// member returns the raft node, or a dsserr.Unavailable error if this node is not a member of the group.
func (s *nodeState) member() (raft.Node, error) {
	node, state := s.get()
	if state != NodeMember {
		return nil, stacktrace.NewErrorWithCode(dsserr.Unavailable, "this node is not a member of the Raft group, its state is %s", state)
	}
	return node, nil
}

// start makes this node a member of the group with the raft node returned by newNode, if this node is waiting.
// It returns whether the raft node was started, and the state of this node.
func (s *nodeState) start(newNode func() raft.Node) (bool, NodeState) {
	s.Lock()
	defer s.Unlock()
	if s.state != NodeWaiting {
		return false, s.state
	}
	s.node = newNode()
	s.state = NodeMember
	return true, s.state
}

// setRemoved records that this node was removed from the group.
func (s *nodeState) setRemoved() {
	s.Lock()
	defer s.Unlock()
	s.state = NodeRemoved
}
