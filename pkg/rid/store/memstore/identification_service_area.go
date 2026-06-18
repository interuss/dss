package memstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
)

func (r *repo) GetISA(_ context.Context, id dssmodels.ID, forUpdate bool) (*ridmodels.IdentificationServiceArea, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetISA not implemented for memstore")
}

func (r *repo) DeleteISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteISA not implemented for memstore")
}

func (r *repo) InsertISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "InsertISA not implemented for memstore")
}

func (r *repo) UpdateISA(_ context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "UpdateISA not implemented for memstore")
}

func (r *repo) SearchISAs(_ context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "SearchISAs not implemented for memstore")
}

func (r *repo) ListExpiredISAs(_ context.Context, writer string, threshold time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "ListExpiredISAs not implemented for memstore")
}

func (r *repo) CountISAs(_ context.Context) (int64, error) {
	return 0, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "CountISAs not implemented for memstore")
}
