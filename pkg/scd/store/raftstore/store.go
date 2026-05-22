package raftstore

import (
	"context"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	"github.com/interuss/dss/pkg/scd/repos"
	scdraftparams "github.com/interuss/dss/pkg/scd/store/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of scd.repos.Repository for Raft-based storage.
type repo struct{}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	params, err := scdraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get scd raft parameters")
	}
	return raftstore.Init(ctx, logger, params, "scd", &repo{})
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) IsReadOnly(_ raftstore.RequestType) bool { return false }

func (r *repo) GetSnapshot() ([]byte, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "not implemented yet")
}

func (r *repo) RestoreFromSnapshot([]byte) error {
	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "not implemented yet")
}

func (r *repo) Apply(_ context.Context, _ consensus.Proposal) (any, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "not implemented yet")
}
