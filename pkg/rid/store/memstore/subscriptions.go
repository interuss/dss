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

func (r *repo) GetSubscription(_ context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "GetSubscription not implemented for memstore")
}

func (r *repo) DeleteSubscription(_ context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "DeleteSubscription not implemented for memstore")
}

func (r *repo) InsertSubscription(_ context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "InsertSubscription not implemented for memstore")
}

func (r *repo) UpdateSubscription(_ context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "UpdateSubscription not implemented for memstore")
}

func (r *repo) SearchSubscriptions(_ context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "SearchSubscriptions not implemented for memstore")
}

func (r *repo) SearchSubscriptionsByOwner(_ context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "SearchSubscriptionsByOwner not implemented for memstore")
}

func (r *repo) UpdateNotificationIdxsInCells(_ context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "UpdateNotificationIdxsInCells not implemented for memstore")
}

func (r *repo) MaxSubscriptionCountInCellsByOwner(_ context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	return 0, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "MaxSubscriptionCountInCellsByOwner not implemented for memstore")
}

func (r *repo) ListExpiredSubscriptions(_ context.Context, writer string, threshold time.Time) ([]*ridmodels.Subscription, error) {
	return nil, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "ListExpiredSubscriptions not implemented for memstore")
}

func (r *repo) CountSubscriptions(_ context.Context) (int64, error) {
	return 0, stacktrace.NewErrorWithCode(dsserr.NotImplemented, "CountSubscriptions not implemented for memstore")
}
