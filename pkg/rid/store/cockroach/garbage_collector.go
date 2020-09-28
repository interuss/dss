package cockroach

import (
	"context"

	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/dss/pkg/rid/repos"
	"go.uber.org/zap"
)

type GarbageCollector struct {
	repos  repos.Repository
	writer string
}

func NewGarbageCollector(repos repos.Repository, writer string) *GarbageCollector {
	return &GarbageCollector{
		repos:  repos,
		writer: writer,
	}
}

func (gc *GarbageCollector) DeleteExpiredRecords(ctx context.Context) {
	gc.DeleteExpiredISAs(ctx)
	gc.DeleteExpiredSubscriptions(ctx)
}

func (gc *GarbageCollector) DeleteExpiredISAs(ctx context.Context) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)
	expiredISAs, err := gc.repos.ListExpiredISAs(ctx, gc.writer)
	if err != nil {
		logger.Panic("Failed to list expired ISAs", zap.Error(err))
		return
	}

	for _, isa := range expiredISAs {
		saOut, err := gc.repos.DeleteISA(ctx, isa)
		if saOut != nil {
			logger.Info("Deleted ISA", zap.Any("ISA", saOut))
		}
		if err != nil {
			logger.Panic("Failed to delete ISAs", zap.Error(err))
		}
	}
}

func (gc *GarbageCollector) DeleteExpiredSubscriptions(ctx context.Context) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)
	expiredSubscriptions, err := gc.repos.ListExpiredSubscriptions(ctx, gc.writer)
	if err != nil {
		logger.Panic("Failed to list expired Subscriptions", zap.Error(err))
		return
	}

	for _, sub := range expiredSubscriptions {
		subOut, err := gc.repos.DeleteSubscription(ctx, sub)
		if subOut != nil {
			logger.Info("Deleted Subscription", zap.Any("Subscription", subOut))
		}
		if err != nil {
			logger.Panic("Failed to delete Subscription", zap.Error(err))
		}
	}
}
