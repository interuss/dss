package consensus

import (
	"sync"

	"github.com/google/uuid"
)

// pendingRead tracks a ReadIndex request
// once the matching raft.ReadState arrives, index holds the commit index the caller must wait
// for its own appliedIndex to reach before it is safe to read local state.
type pendingRead struct {
	index      uint64
	indexKnown bool
	done       chan struct{}
}

// readIndexTracker keeps track of pending ReadIndex requests
type readIndexTracker struct {
	sync.Mutex
	pending map[uuid.UUID]*pendingRead
}

func newReadIndexTracker() *readIndexTracker {
	return &readIndexTracker{
		pending: make(map[uuid.UUID]*pendingRead),
	}
}

// track registers a pending read and returns a channel that closes once it is safe to read local state
func (r *readIndexTracker) track(id uuid.UUID) <-chan struct{} {
	r.Lock()
	defer r.Unlock()

	p := &pendingRead{done: make(chan struct{})}
	r.pending[id] = p
	return p.done
}

// setIndex records the commit index a ReadState reported for id, and
// immediately releases the waiter if appliedIndex already satisfies it.
func (r *readIndexTracker) setIndex(id uuid.UUID, index uint64, appliedIndex uint64) {
	r.Lock()
	defer r.Unlock()

	p, ok := r.pending[id]
	if !ok {
		return
	}

	p.index = index
	p.indexKnown = true

	if index <= appliedIndex {
		close(p.done)
		delete(r.pending, id)
	}
}

// releaseUpTo signals and removes every pending read whose recorded index is now satisfied by appliedIndex.
// Must be called after every appliedIndex advance, since a read's ReadState may have arrived before the entries it depends on were applied.
func (r *readIndexTracker) releaseUpTo(appliedIndex uint64) {
	r.Lock()
	defer r.Unlock()

	for id, p := range r.pending {
		if p.indexKnown && p.index <= appliedIndex {
			close(p.done)
			delete(r.pending, id)
		}
	}
}

// untrack removes a pending read without signaling it
// e.g. when the caller's context is done before a matching ReadState ever arrived.
func (r *readIndexTracker) untrack(id uuid.UUID) {
	r.Lock()
	defer r.Unlock()

	delete(r.pending, id)
}
