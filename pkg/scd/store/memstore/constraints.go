package memstore

import (
	"context"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) SearchConstraints(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.Constraint, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "SearchConstraints not implemented for memstore")
}

func (r *repo) GetConstraint(_ context.Context, id dssmodels.ID) (*scdmodels.Constraint, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetConstraint not implemented for memstore")
}

func (r *repo) UpsertConstraint(_ context.Context, constraint *scdmodels.Constraint) (*scdmodels.Constraint, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "UpsertConstraint not implemented for memstore")
}

func (r *repo) DeleteConstraint(_ context.Context, id dssmodels.ID) error {
	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteConstraint not implemented for memstore")
}

func (r *repo) CountConstraints(_ context.Context) (int64, error) {
	return 0, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "CountConstraint not implemented for memstore")
}
