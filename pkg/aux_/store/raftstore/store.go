package raftstore

import (
	"github.com/interuss/dss/pkg/aux_/repos"
	"github.com/interuss/dss/pkg/raftstore"
)

// repo is a full implementation of aux_.repos.Repository for Raft-based storage.
type repo struct{}

func Init() (*raftstore.Store[repos.Repository], error) {
	return raftstore.Init[repos.Repository](func() repos.Repository { return &repo{} }), nil
}
