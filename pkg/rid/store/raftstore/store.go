package raftstore

import (
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/rid/repos"
	"go.uber.org/zap"
)

// repo is a full implementation of rid.repos.Repository for Raft-based storage.
type repo struct{}

func Init(logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	return raftstore.Init[repos.Repository](logger, func() repos.Repository { return &repo{} })
}
