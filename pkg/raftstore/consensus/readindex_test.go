package consensus

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

const defaultReleaseTimeout = 100 * time.Millisecond

func assertReleased(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(defaultReleaseTimeout):
		t.Fatal("expected pending read to be released")
	}
}

func assertNotReleased(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
		t.Fatal("expected pending read to still be pending")
	case <-time.After(defaultReleaseTimeout):
	}
}

func TestReadIndexTracker_AppliedIndexAlreadyCaughtUp(t *testing.T) {
	r := newReadIndexTracker()
	id := uuid.New()
	done := r.track(id)

	r.setIndex(id, 5, 10)

	assertReleased(t, done)
}

func TestReadIndexTracker_AppliedIndexCatchesUpLater(t *testing.T) {
	r := newReadIndexTracker()
	id := uuid.New()
	done := r.track(id)

	r.setIndex(id, 10, 5)
	assertNotReleased(t, done)

	r.releaseUpTo(9)
	assertNotReleased(t, done)

	r.releaseUpTo(10)
	assertReleased(t, done)
}

func TestReadIndexTracker_MultiplePendingReadsReleaseIndependently(t *testing.T) {
	r := newReadIndexTracker()
	idA, idB := uuid.New(), uuid.New()
	doneA := r.track(idA)
	doneB := r.track(idB)

	r.setIndex(idA, 3, 0)
	r.setIndex(idB, 8, 0)

	r.releaseUpTo(3)
	assertReleased(t, doneA)
	assertNotReleased(t, doneB)

	r.releaseUpTo(8)
	assertReleased(t, doneB)
}
