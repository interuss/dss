package raftstore

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
)

func (r *repo) SearchConstraints(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Constraint, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) GetConstraint(_ context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) UpsertConstraint(_ context.Context, constraint *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) DeleteConstraint(_ context.Context, id dssmodels.ID) error {
	// TODO: implement
	return nil
}
