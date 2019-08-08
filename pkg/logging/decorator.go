package logging

// import (
// 	"context"

// 	"github.com/golang/geo/s2"
// 	"github.com/steeling/InterUSS-Platform/pkg/dss"
// 	dspb "github.com/steeling/InterUSS-Platform/pkg/dssproto"
// 	"go.uber.org/zap"
// )

// type loggingStore struct {
// 	logger *zap.Logger
// 	next   dss.Store
// }

// func (ls *loggingStore) Close() error {
// 	err := ls.next.Close()
// 	ls.logger.Debug("Store.Close", zap.Error(err))
// 	return err
// }

// func (ls *loggingStore) DeleteIdentificationServiceArea(ctx context.Context, id string, owner string) (*dspb.IdentificationServiceArea, []*dspb.SubscriberToNotify, error) {
// 	area, subscriptions, err := ls.next.DeleteIdentificationServiceArea(ctx, id, owner)
// 	ls.logger.Debug(
// 		"Store.DeleteIdentificationServiceArea",
// 		zap.String("id", id),
// 		zap.String("owner", owner),
// 		zap.Any("area", area),
// 		zap.Any("subscriptions", subscriptions),
// 		zap.Error(err),
// 	)
// 	return area, subscriptions, err
// }
// func (ls *loggingStore) GetSubscription(ctx context.Context, id string) (*dspb.Subscription, error) {
// 	subscription, err := ls.next.GetSubscription(ctx, id)
// 	ls.logger.Debug("Store.GetSubscription", zap.String("id", id), zap.Any("subscription", subscription), zap.Error(err))
// 	return subscription, err
// }

// func (ls *loggingStore) DeleteSubscription(ctx context.Context, id, version string) (*dspb.Subscription, error) {
// 	subscription, err := ls.next.DeleteSubscription(ctx, id, version)
// 	ls.logger.Debug("Store.DeleteSubscription", zap.String("id", id), zap.String("version", version), zap.Any("subscription", subscription), zap.Error(err))
// 	return subscription, err
// }

// func (ls *loggingStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion, owner string) ([]*dspb.Subscription, error) {
// 	subscriptions, err := ls.next.SearchSubscriptions(ctx, cells, owner)
// 	ls.logger.Debug("Store.SearchSubscriptions", zap.Any("cells", cells), zap.String("owner", owner), zap.Any("subscriptions", subscriptions), zap.Error(err))
// 	return subscriptions, err
// }
