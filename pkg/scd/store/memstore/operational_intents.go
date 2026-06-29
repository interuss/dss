package memstore

import (
	"context"
	"time"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) GetOperationalIntent(_ context.Context, id dssmodels.ID) (*scdmodels.OperationalIntent, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetOperationalIntent not implemented for memstore")
}

func (r *repo) DeleteOperationalIntent(_ context.Context, id dssmodels.ID) error {
	return stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteOperationalIntent not implemented for memstore")
}

func (r *repo) UpsertOperationalIntent(_ context.Context, operation *scdmodels.OperationalIntent) (*scdmodels.OperationalIntent, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "UpsertOperationalIntent not implemented for memstore")
}

func (r *repo) SearchOperationalIntents(_ context.Context, v4d *dssmodels.Volume4D) ([]*scdmodels.OperationalIntent, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "SearchOperationalIntents not implemented for memstore")
}

func (r *repo) GetDependentOperationalIntents(_ context.Context, subscriptionID dssmodels.ID) ([]dssmodels.ID, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetDependentOperationalIntents not implemented for memstore")
}

func (r *repo) ListExpiredOperationalIntents(_ context.Context, threshold time.Time) ([]*scdmodels.OperationalIntent, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "ListExpiredOperationalIntents not implemented for memstore")
}

func (r *repo) CountOperationalIntents(_ context.Context) (int64, error) {
	return 0, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "CountOperationalIntents not implemented for memstore")
}
