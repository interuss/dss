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
	node  raft.Node
}

// get returns the raft node and the state of this node.
func (s *nodeState) get() (raft.Node, NodeState) {
	s.RLock()
	defer s.RUnlock()
	return s.node, s.state
}

// raftNode returns the raft node.
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

// setRemoved records that this node was removed from the group.
func (s *nodeState) setRemoved() {
	s.Lock()
	defer s.Unlock()
	s.state = NodeRemoved
}
