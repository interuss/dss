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

// StartEviction runs the scheduled cleanup jobs of the group in the background, until ctx is done.
// While this node is the leader, the jobs run according to the eviction settings of repo.
// The schedule restarts when this node becomes the leader or when the settings are updated (configC is signaled) .
// No job runs while this node is not the leader or the interval is 0.
func (s *Store[R]) StartEviction(ctx context.Context, repo storerepos.Eviction, configC <-chan struct{}) {
	logger := s.logger.With(zap.String("component", "eviction"))
	go func() {
		var leader bool
		var since time.Time
		var tick <-chan time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case leader = <-s.LeaderC():
				since = time.Now()
				tick = nextEviction(ctx, logger, repo, leader, since)
			case <-configC:
				tick = nextEviction(ctx, logger, repo, leader, since)
			case <-tick:
				ctx := timestamp.NewContext(ctx, time.Now())
				configCtx, cancel := context.WithTimeout(ctx, evictionConfigTimeout)
				config, err := repo.GetEvictionConfig(configCtx)
				cancel()
				if err != nil {
					logger.Error("failed to get the eviction settings, retrying", zap.Error(err))
					tick = time.After(evictionRetryDelay)
					continue
				}
				counts, err := repo.DeleteExpired(ctx, config.Parameters)
				if err != nil {
					logger.Error("failed to run the scheduled cleanup job", zap.Error(err))
				} else {
					logger.Info("ran scheduled cleanup job", zap.Any("counts", counts))
				}
				since = time.Now()
				tick = nextEviction(ctx, logger, repo, leader, since)
			}
		}
	}()
}

// nextEviction returns a channel receiving when the next cleanup job is due, one interval after since,
// or nil if this node is not the leader or the interval is 0.
// If the settings can't be read, the timer is set to evictionRetryDelay.
func nextEviction(ctx context.Context, logger *zap.Logger, repo storerepos.Eviction, leader bool, since time.Time) <-chan time.Time {
	if !leader {
		return nil
	}
	ctx, cancel := context.WithTimeout(timestamp.NewContext(ctx, time.Now()), evictionConfigTimeout)
	defer cancel()
	config, err := repo.GetEvictionConfig(ctx)
	if err != nil {
		logger.Error("failed to get the eviction settings, retrying", zap.Error(err))
		return time.After(evictionRetryDelay)
	}
	if config.Interval == 0 {
		return nil
	}
	return time.After(time.Until(since.Add(config.Interval)))
}
