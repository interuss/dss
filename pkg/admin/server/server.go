package server

import (
	"github.com/interuss/dss/pkg/raftstore/consensus"
	"go.uber.org/zap"
)

// Server implements adminv1.Implementation.
type Server struct {
	Logger *zap.Logger
	// raftGroups are the Raft groups of the DSS node, indexed by name.
	raftGroups map[string]*consensus.Consensus
}
