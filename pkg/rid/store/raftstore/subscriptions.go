package raftstore

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

func (r *repo) GetSubscription(_ context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) DeleteSubscription(_ context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) InsertSubscription(_ context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) UpdateSubscription(_ context.Context, sub *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) SearchSubscriptions(_ context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) SearchSubscriptionsByOwner(_ context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) UpdateNotificationIdxsInCells(_ context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}

func (r *repo) MaxSubscriptionCountInCellsByOwner(_ context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	// TODO: implement
	return 0, nil
}

func (r *repo) ListExpiredSubscriptions(_ context.Context, writer string, threshold time.Time) ([]*ridmodels.Subscription, error) {
	// TODO: implement
	return nil, nil
}
