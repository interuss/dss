package raftstore

import (
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/scd/repos"
)

// repo is a full implementation of scd.repos.Repository for Raft-based storage.
type repo struct{}

func Init() (*raftstore.Store[repos.Repository], error) {
	return raftstore.Init[repos.Repository](func() repos.Repository { return &repo{} }), nil
}
