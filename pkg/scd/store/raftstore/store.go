package raftstore

import (
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/scd/repos"
	"go.uber.org/zap"
)

// repo is a full implementation of scd.repos.Repository for Raft-based in-memory storage.
type repo struct {
	store *Store
}

func Init(logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	return raftstore.Init
}

// Store implements store.Store[repos.Repository] for Raft-based in-memory storage.
type Store struct {
	logger *zap.Logger
}
