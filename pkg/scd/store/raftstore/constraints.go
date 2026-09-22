package raftstore

import (
	"context"
	"encoding/json"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

const (
	searchConstraints consensus.RequestType[[]*scdmodels.Constraint] = "searchConstraints"
	getConstraint     consensus.RequestType[*scdmodels.Constraint]   = "getConstraint"
	upsertConstraint  consensus.RequestType[*scdmodels.Constraint]   = "upsertConstraint"
	deleteConstraint  consensus.RequestType[any]                     = "deleteConstraint"
	countConstraints  consensus.RequestType[int64]                   = "countConstraints"
)

func (r *repo) SearchConstraints(ctx context.Context, cellsVolume *dssmodels.CellsVolume4D) ([]*scdmodels.Constraint, error) {
	buf, err := json.Marshal(cellsVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, searchConstraints, buf, true)
}

func (r *repo) GetConstraint(ctx context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, getConstraint, buf, true)
}

func (r *repo) UpsertConstraint(ctx context.Context, constraint *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	buf, err := json.Marshal(constraint)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, upsertConstraint, buf, false)
}

func (r *repo) DeleteConstraint(ctx context.Context, id dssmodels.ID) error {
	buf, err := json.Marshal(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal payload")
	}

	_, err = r.consensus.HandleClientRequest(ctx, deleteConstraint, buf, false)
	return err
}

func (r *repo) CountConstraints(ctx context.Context) (int64, error) {
	return r.consensus.HandleClientRequest(ctx, countConstraints, nil, true)
}

func (r *repo) applyConstraint(ctx context.Context, proposal consensus.Proposal) (any, error) {
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
		return nil, stacktrace.NewError("unrecognized constraint request type: %s", proposal.RequestType)
	}
}
