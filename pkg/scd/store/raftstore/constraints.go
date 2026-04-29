package raftstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

func (r *repo) SearchConstraints(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Constraint, error) {
	panic("SearchConstraints not yet implemented in raft store")
}

func (r *repo) GetConstraint(_ context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
	panic("GetConstraint not yet implemented in raft store")
}

func (r *repo) UpsertConstraint(_ context.Context, constraint *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	panic("UpsertConstraint not yet implemented in raft store")
}

func (r *repo) DeleteConstraint(_ context.Context, id dssmodels.ID) error {
	panic("DeleteConstraint not yet implemented in raft store")
}
