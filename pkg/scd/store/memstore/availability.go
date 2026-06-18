package memstore

import (
	"context"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) GetUssAvailability(_ context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetUssAvailability not implemented for memstore")
}

func (r *repo) UpsertUssAvailability(_ context.Context, ussa *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "UpsertUssAvailability not implemented for memstore")
}
