package consensus

import (
	"context"
	"sync"
)

// leaderRoutines runs routines while this node is the leader of the group.
type leaderRoutines struct {
	sync.Mutex
	routines []leaderRoutine
	cancel   context.CancelFunc // cancels the running routines, nil while this node is not the leader
}

type leaderRoutine struct {
	ctx context.Context // parent context of each run of fn
	fn  func(ctx context.Context)
}

// add registers fn and starts it if this node is the leader.
func (l *leaderRoutines) add(ctx context.Context, fn func(ctx context.Context)) {
	l.Lock()
	defer l.Unlock()
	r := leaderRoutine{ctx: ctx, fn: fn}
	l.routines = append(l.routines, r)
	if l.cancel != nil {
		l.start(r)
	}
}

// set starts the routines when this node becomes the leader and cancels them when it stops being one.
func (l *leaderRoutines) set(leader bool) {
	l.Lock()
	defer l.Unlock()
	if leader == (l.cancel != nil) {
		return
	}
	if !leader {
		l.cancel()
		l.cancel = nil
		return
	}
	l.cancel = func() {}
	for _, r := range l.routines {
		l.start(r)
	}
}

func (l *leaderRoutines) start(r leaderRoutine) {
	ctx, cancel := context.WithCancel(r.ctx)
	cancelOthers := l.cancel
	l.cancel = func() { cancelOthers(); cancel() }
	go r.fn(ctx)
}
