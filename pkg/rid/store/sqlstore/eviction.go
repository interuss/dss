package sqlstore

import (
	"context"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) DeleteExpired(_ context.Context, _ dssmodels.EvictionParameters) (map[string]int, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteExpired not implemented for sqlstore")
}

func (r *repo) GetEvictionState(_ context.Context) (dssmodels.EvictionState, error) {
	return dssmodels.EvictionState{}, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetEvictionState not implemented for sqlstore")
}

func (r *repo) SetEvictionConfig(_ context.Context, _ dssmodels.EvictionConfig) (dssmodels.EvictionConfig, error) {
	return dssmodels.EvictionConfig{}, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "SetEvictionConfig not implemented for sqlstore")
}
