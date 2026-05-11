package raftstore

import (
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/rid/repos"
)

// repo is a full implementation of rid.repos.Repository for Raft-based storage.
type repo struct{}

func Init() (*raftstore.Store[repos.Repository], error) {
	return raftstore.Init[repos.Repository](func() repos.Repository { return &repo{} }), nil
}
