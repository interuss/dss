package raftstore

import (
	"context"
	"time"

	storerepos "github.com/interuss/dss/pkg/store/repos"
	"github.com/interuss/dss/pkg/timestamp"
	"go.uber.org/zap"
)

const (
	evictionRetryDelay    = 10 * time.Second
	evictionConfigTimeout = 10 * time.Second
)

// StartEviction runs the scheduled cleanup jobs of the group while this node is the leader, according to the eviction settings of repo.
// The schedule is computed from the last cleanup job of the group when this node becomes the leader
// and recomputed when the settings or the last cleanup job are updated (configC is signaled).
// No job runs while the interval is 0.
func (s *Store[R]) StartEviction(ctx context.Context, repo storerepos.Eviction, configC <-chan struct{}) {
	logger := s.logger.With(zap.String("component", "eviction"))
	s.RunAsLeader(ctx, func(ctx context.Context) {
		tick := nextEviction(ctx, logger, repo)
		for {
			select {
			case <-ctx.Done():
				return
			case <-configC:
				tick = nextEviction(ctx, logger, repo)
			case <-tick:
				ctx := timestamp.NewContext(ctx, time.Now())
				configCtx, cancel := context.WithTimeout(ctx, evictionConfigTimeout)
				state, err := repo.GetEvictionState(configCtx)
				cancel()
				if err != nil {
					logger.Error("failed to get the eviction settings, retrying", zap.Error(err))
					tick = time.After(evictionRetryDelay)
					continue
				}
				counts, err := repo.DeleteExpired(ctx, state.Config.Parameters)
				if err != nil {
					logger.Error("failed to run the scheduled cleanup job, retrying", zap.Error(err))
					tick = time.After(evictionRetryDelay)
					continue
				}
				logger.Info("ran scheduled cleanup job", zap.Any("counts", counts))
				tick = nextEviction(ctx, logger, repo)
			}
		}
	})
}

// nextEviction returns a channel receiving when the next cleanup job is due, one interval after the last one or
// immediately if none ran, or nil if the interval is 0.
// If the settings can't be read, the timer is set to evictionRetryDelay.
func nextEviction(ctx context.Context, logger *zap.Logger, repo storerepos.Eviction) <-chan time.Time {
	ctx, cancel := context.WithTimeout(timestamp.NewContext(ctx, time.Now()), evictionConfigTimeout)
	defer cancel()
	state, err := repo.GetEvictionState(ctx)
	if err != nil {
		logger.Error("failed to get the eviction settings, retrying", zap.Error(err))
		return time.After(evictionRetryDelay)
	}
	if state.Config.Interval == 0 {
		return nil
	}
	return time.After(time.Until(state.LastRun.Add(state.Config.Interval)))
}
