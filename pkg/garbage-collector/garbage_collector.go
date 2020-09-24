package gc

import (
	"context"

	"go.uber.org/zap"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/rid/repos"
)

type GarbageCollector struct {
	isaRepo repos.ISA
	subscriptionRepo repos.Subscription
	writer string
}

func(gc *GarbageCollector) DeleteExpiredRecords(ctx context.Context) {
	gc.DeleteExpiredISAs(ctx)
	gc.DeleteExpiredSubscriptions(ctx)
}

func(gc *GarbageCollector) DeleteExpiredISAs(ctx context.Context) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)
	expiredISAs, err := gc.isaRepo.ListExpiredISAs(ctx, gc.writer)
	if err != nil {
		logger.Panic("Failed to list expired ISAs", zap.Error(err))
		return
	}

	for _, isa := range expiredISAs {
		saOut, err := gc.isaRepo.DeleteISA(ctx, isa)
		if saOut != nil {
			logger.Info("Deleted ISA", zap.Any("ISA", saOut))
		}
		if err != nil {
			logger.Panic("Failed to delete ISAs", zap.Error(err))
		}
	}
}

func(gc *GarbageCollector) DeleteExpiredSubscriptions(ctx context.Context) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)
	expiredSubscriptions, err := gc.subscriptionRepo.ListExpiredSubscriptions(ctx, gc.writer)
	if err != nil {
		logger.Panic("Failed to list expired Subscriptions", zap.Error(err))
		return
	}

	for _, sub := range expiredSubscriptions {
		subOut, err := gc.subscriptionRepo.DeleteSubscription(ctx, sub)
		if subOut != nil {
			logger.Info("Deleted Subscription", zap.Any("Subscription", subOut))
		}
		if err != nil {
			logger.Panic("Failed to delete ISA", zap.Error(err))
		}
	}
}