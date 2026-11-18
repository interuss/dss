package consensus

import (
	"context"
	"sync"
)

// leaderRoutines runs routines while this node is the leaderRoutines of the group.
type leaderRoutines struct {
	sync.Mutex
	routines []func(ctx context.Context)
	cancels  []context.CancelFunc // cancel the running routines
	isLeader bool
}

// add registers fn and starts it if this node is the leader.
func (l *leaderRoutines) add(fn func(ctx context.Context)) {
	l.Lock()
	defer l.Unlock()
	l.routines = append(l.routines, fn)
	if l.isLeader {
		l.start(fn)
	}
}

// set starts the routines when this node becomes the leader and cancels them when it stops being one.
func (l *leaderRoutines) set(leader bool) {
	l.Lock()
	defer l.Unlock()
	if leader == l.isLeader {
		return
	}
	l.isLeader = leader
	if !leader {
		for _, cancel := range l.cancels {
			cancel()
		}
		l.cancels = nil
		return
	}
	for _, fn := range l.routines {
		l.start(fn)
	}
}

func (l *leaderRoutines) start(fn func(ctx context.Context)) {
	ctx, cancel := context.WithCancel(context.Background())
	l.cancels = append(l.cancels, cancel)
	go fn(ctx)
}
