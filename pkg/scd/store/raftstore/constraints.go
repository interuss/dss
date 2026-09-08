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
	searchConstraints consensus.RequestType = "searchConstraints"
	getConstraint     consensus.RequestType = "getConstraint"
	upsertConstraint  consensus.RequestType = "upsertConstraint"
	deleteConstraint  consensus.RequestType = "deleteConstraint"
	countConstraints  consensus.RequestType = "countConstraints"
)

func (r *repo) SearchConstraints(ctx context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Constraint, error) {
	buf, err := json.Marshal(v4d)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, searchConstraints, buf, true)
	if err != nil {
		return nil, err
	}
	if constraints, ok := result.([]*scdmodels.Constraint); ok {
		return constraints, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) GetConstraint(ctx context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, getConstraint, buf, true)
	if err != nil {
		return nil, err
	}
	if constraint, ok := result.(*scdmodels.Constraint); ok {
		return constraint, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) UpsertConstraint(ctx context.Context, constraint *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	buf, err := json.Marshal(constraint)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	result, err := r.consensus.HandleClientRequest(ctx, upsertConstraint, buf, false)
	if err != nil {
		return nil, err
	}
	if upserted, ok := result.(*scdmodels.Constraint); ok {
		return upserted, nil
	}
	return nil, stacktrace.NewError("unexpected result type: %T", result)
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
	result, err := r.consensus.HandleClientRequest(ctx, countConstraints, nil, true)
	if err != nil {
		return 0, err
	}
	if count, ok := result.(int64); ok {
		return count, nil
	}
	return 0, stacktrace.NewError("unexpected result type: %T", result)
}

func (r *repo) applyConstraint(ctx context.Context, proposal consensus.Proposal) (any, error) {
	switch proposal.RequestType {
	case searchConstraints:
		var v4d dssmodels.Volume4D
		if err := json.Unmarshal(proposal.Value, &v4d); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchConstraints)
		}
		return r.Store.GetRepo().SearchConstraints(ctx, &v4d)

	case getConstraint:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getConstraint)
		}
		return r.Store.GetRepo().GetConstraint(ctx, id)

	case upsertConstraint:
		var constraint scdmodels.Constraint
		if err := json.Unmarshal(proposal.Value, &constraint); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", upsertConstraint)
		}
		return r.Store.GetRepo().UpsertConstraint(ctx, &constraint)

	case deleteConstraint:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteConstraint)
		}
		return nil, r.Store.GetRepo().DeleteConstraint(ctx, id)

	case countConstraints:
		return r.Store.GetRepo().CountConstraints(ctx)

	default:
		return nil, stacktrace.NewError("unrecognized constraint request type: %s", proposal.RequestType)
	}
}
