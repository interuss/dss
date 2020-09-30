package cockroach

import (
	"context"

	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/stacktrace"
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

func (gc *GarbageCollector) DeleteRIDExpiredRecords(ctx context.Context) error {
	err := gc.DeleteExpiredISAs(ctx)
	if err != nil {
		return stacktrace.Propagate(err,
			"Failed to delete RID expired records")
	}
	err = gc.DeleteExpiredSubscriptions(ctx)
	if err != nil {
		return stacktrace.Propagate(err,
			"Failed to delete RID expired records")
	}

	return nil
}

func (gc *GarbageCollector) DeleteExpiredISAs(ctx context.Context) error {
	expiredISAs, err := gc.repos.ListExpiredISAs(ctx, gc.writer)
	if err != nil {
		return stacktrace.Propagate(err,
			"Failed to list expired ISAs")
	}

	for _, isa := range expiredISAs {
		isaOut, err := gc.repos.DeleteISA(ctx, isa)
		if isaOut != nil {
			return stacktrace.Propagate(err,
				"Deleted ISA")
		}
		if err != nil {
			return stacktrace.Propagate(err,
				"Failed to delete ISAs")
		}
	}

	return nil
}

func (gc *GarbageCollector) DeleteExpiredSubscriptions(ctx context.Context) error {
	expiredSubscriptions, err := gc.repos.ListExpiredSubscriptions(ctx, gc.writer)
	if err != nil {
		return stacktrace.Propagate(err,
			"Failed to list expired Subscriptions")
	}

	for _, sub := range expiredSubscriptions {
		subOut, err := gc.repos.DeleteSubscription(ctx, sub)
		if subOut != nil {
			return stacktrace.Propagate(err,
				"Deleted Subscription")
		}
		if err != nil {
			return stacktrace.Propagate(err,
				"Failed to delete Subscription")
		}
	}
	return nil
}
