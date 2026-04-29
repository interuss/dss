package raftstore

import (
	"context"

	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"go.uber.org/zap"
)

// Store implements store.Store[repos.Repository] for Raft-based in-memory storage.
type Store struct {
	logger *zap.Logger
}

// repo is a full implementation of scd.repos.Repository for Raft-based in-memory storage.
type repo struct {
	store *Store
}

func Init(logger *zap.Logger) (dssstore.Store[repos.Repository], error) {
	return &Store{logger: logger}, nil
}

func (s *Store) Close() error {
	panic("Close not yet implemented in raft store")
}

func (s *Store) Interact(_ context.Context) (repos.Repository, error) {
	panic("Interact not yet implemented in raft store")
}

func (s *Store) Transact(_ context.Context, _ func(context.Context, repos.Repository) error) error {
	panic("Transact not yet implemented in raft store")
}
