package server

import (
	"github.com/interuss/dss/pkg/admin"
	"go.uber.org/zap"
)

// Server implements adminv1.Implementation.
type Server struct {
	Logger *zap.Logger
	// RaftGroups are the Raft groups of the DSS node, indexed by name.
	RaftGroups map[string]admin.RaftGroup
}
