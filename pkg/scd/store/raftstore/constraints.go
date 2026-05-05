package raftstore

import (
	"context"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) SearchConstraints(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Constraint, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "SearchConstraints not yet implemented in raft store")
}

func (r *repo) GetConstraint(_ context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetConstraint not yet implemented in raft store")
}

func (r *repo) UpsertConstraint(_ context.Context, constraint *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "UpsertConstraint not yet implemented in raft store")
}

func (r *repo) DeleteConstraint(_ context.Context, id dssmodels.ID) error {
	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteConstraint not yet implemented in raft store")
}
