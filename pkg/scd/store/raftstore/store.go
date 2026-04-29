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
	s.logger.Warn("raft store Close not yet implemented")
	return nil
}

func (s *Store) Interact(_ context.Context) (repos.Repository, error) {
	s.logger.Warn("raft store Interact not yet implemented")
	return &repo{store: s}, nil
}

func (s *Store) Transact(ctx context.Context, f func(context.Context, repos.Repository) error) error {
	s.logger.Warn("raft store Transact not yet implemented")
	return f(ctx, &repo{store: s})
}
