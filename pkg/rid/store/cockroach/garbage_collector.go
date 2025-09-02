package cockroach

import (
	"context"

	"github.com/interuss/dss/pkg/rid/repos"
	"github.com/interuss/stacktrace"
)

type GarbageCollector struct {
	store  *Store
	writer string
}

func NewGarbageCollector(store *Store, writer string) *GarbageCollector {
	return &GarbageCollector{
		store:  store,
		writer: writer,
	}
}

func (gc *GarbageCollector) DeleteRIDExpiredRecords(ctx context.Context) error {
	repos, err := gc.store.Interact(ctx)
	if err != nil {
		return stacktrace.Propagate(err,
			"Unable to interact with store")
	}

	err = gc.DeleteExpiredISAs(ctx, repos)
	if err != nil {
		return stacktrace.Propagate(err,
			"Failed to delete RID expired records")
	}

	err = gc.DeleteExpiredSubscriptions(ctx, repos)
	if err != nil {
		return stacktrace.Propagate(err,
			"Failed to delete RID expired records")
	}

	return nil
}

func (gc *GarbageCollector) DeleteExpiredISAs(ctx context.Context, repos repos.Repository) error {
	expiredISAs, err := repos.ListExpiredISAs(ctx, gc.writer)
	if err != nil {
		return stacktrace.Propagate(err,
			"Failed to list expired ISAs")
	}

	for _, isa := range expiredISAs {
		isaOut, err := repos.DeleteISA(ctx, isa)
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

func (gc *GarbageCollector) DeleteExpiredSubscriptions(ctx context.Context, repos repos.Repository) error {
	expiredSubscriptions, err := repos.ListExpiredSubscriptions(ctx, gc.writer)
	if err != nil {
		return stacktrace.Propagate(err,
			"Failed to list expired Subscriptions")
	}

	for _, sub := range expiredSubscriptions {
		subOut, err := repos.DeleteSubscription(ctx, sub)
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
