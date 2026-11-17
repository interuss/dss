package server

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	"go.uber.org/zap"
)

// Eviction is the eviction of a Raft group, implemented by its repository.
type Eviction interface {
	DeleteExpired(ctx context.Context, params dssmodels.EvictionParameters) (map[string]int, error)
	GetEvictionConfig(ctx context.Context) (*dssmodels.EvictionConfig, error)
	SetEvictionConfig(ctx context.Context, config dssmodels.EvictionConfig) (*dssmodels.EvictionConfig, error)
}

// Server implements adminv1.Implementation.
type Server struct {
	Logger *zap.Logger
	// raftGroups are the Raft groups of the DSS node, indexed by name.
	raftGroups map[string]*consensus.Consensus
	// evictions are the evictions of the Raft groups with eviction, indexed by group name.
	evictions map[string]Eviction
}
