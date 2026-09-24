package raftstore

import (
	"context"
	"encoding/json"

	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/operations"
	"github.com/interuss/dss/pkg/scd/repos"
	scdmemstore "github.com/interuss/dss/pkg/scd/store/memstore"
	scdraftparams "github.com/interuss/dss/pkg/scd/store/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

// repo is a full implementation of scd.repos.Repository for Raft-based storage.
type repo struct {
	consensus *consensus.Consensus
	*memstore.Store[repos.Repository]
}

func Init(ctx context.Context, logger *zap.Logger, locality string) (*raftstore.Store[repos.Repository], error) {
	params, err := scdraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get scd raft parameters")
	}

	memStore, err := scdmemstore.Init(ctx, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize scd memstore")
	}

	r := &repo{Store: memStore}
	store, err := raftstore.Init(ctx, logger.With(zap.String("service", "scd")), locality, params, r, operations.Registry)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize scd raftstore")
	}

	r.consensus = store.Consensus

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	case string(searchConstraints):
		var cellsVolume dssmodels.CellsVolume4D
		if err := json.Unmarshal(proposal.Value, &cellsVolume); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchConstraints)
		}
		return r.Store.GetRepo().SearchConstraints(ctx, &cellsVolume)

	case string(getConstraint):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getConstraint)
		}
		return r.Store.GetRepo().GetConstraint(ctx, id)

	case string(upsertConstraint):
		var constraint scdmodels.Constraint
		if err := json.Unmarshal(proposal.Value, &constraint); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", upsertConstraint)
		}
		return r.Store.GetRepo().UpsertConstraint(ctx, &constraint)

	case string(deleteConstraint):
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteConstraint)
		}
		return nil, r.Store.GetRepo().DeleteConstraint(ctx, id)

	case string(countConstraints):
		return r.Store.GetRepo().CountConstraints(ctx)

	default:
		handler, ok := operations.Registry[proposal.RequestType]
		if !ok {
			return nil, stacktrace.NewError("unrecognized request type: %s", proposal.RequestType)
		}

		request, err := handler.Decode(proposal.Value)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to decode %s payload", proposal.RequestType)
		}

		return handler.Execute(ctx, r.Store.GetRepo(), request)
	}
}
