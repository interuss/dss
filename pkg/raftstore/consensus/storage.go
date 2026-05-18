package consensus

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	"go.etcd.io/etcd/server/v3/storage/wal"
	"go.etcd.io/etcd/server/v3/storage/wal/walpb"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
)

const snapshotCatchUpEntriesN uint64 = 10000

// snapshotProvider is a function that returns the snapshot data to be included in the Raft snapshot.
// We use a snapshotProvider from each component (scd, rid and aux).
type snapshotProvider func() ([]byte, error)

// storage persists the Raft log and snapshots and manages the raft in-memory storage.
type storage struct {
	logger *zap.Logger

	*raft.MemoryStorage

	wal *wal.WAL

	snapper   *snap.Snapshotter
	providers map[string]snapshotProvider
}

func newStorage(logger *zap.Logger, dataDir string, nodeID uint64) (*storage, bool, error) {
	// load the latest snapshot
	snapshotPath := path.Join(dataDir, fmt.Sprintf("snapshot_%d", nodeID))
	if !fileutil.Exist(snapshotPath) {
		err := os.MkdirAll(snapshotPath, 0o750)
		if err != nil {
			return nil, false, stacktrace.Propagate(err, "failed to create directory `%s`for snapshot storage", snapshotPath)
		}
	}

	walPath := path.Join(dataDir, fmt.Sprintf("wal_%d", nodeID))

	snapper := snap.New(logger, snapshotPath)
	snapshot, err := loadSnapshot(logger, walPath, snapper)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "failed to load snapshot")
	}

	// load the wal entries
	var w *wal.WAL
	ok := wal.Exist(walPath)
	if !ok {
		err := os.MkdirAll(walPath, 0o750)
		if err != nil {
			return nil, false, stacktrace.Propagate(err, "failed to create directory `%s` for wal storage", walPath)
		}

		w, err := wal.Create(logger, walPath, nil)
		if err != nil {
			return nil, false, stacktrace.Propagate(err, "failed to create wal at %s", walPath)
		}

		err = w.Close()
		if err != nil {
			return nil, false, stacktrace.Propagate(err, "failed to close wal at %s", walPath)
		}
	}

	// open the wal at the given snapshot and get all subsequent entries
	w, err = wal.Open(logger, walPath, walpb.Snapshot{
		Index:     snapshot.Metadata.Index,
		Term:      snapshot.Metadata.Term,
		ConfState: &snapshot.Metadata.ConfState,
	})
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "failed to open wal at %s with index %d, term %d and confstate %v", walPath, snapshot.Metadata.Index, snapshot.Metadata.Term, snapshot.Metadata.ConfState)
	}

	_, state, entries, err := w.ReadAll()
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "failed to read wal state and entries")
	}

	// initialize the raft memory storage with the loaded snapshot and wal entries
	raftMemoryStorage := raft.NewMemoryStorage()
	if !raft.IsEmptySnap(*snapshot) {
		err = raftMemoryStorage.ApplySnapshot(*snapshot)
		if err != nil {
			return nil, false, stacktrace.Propagate(err, "failed to apply snapshot to raft memory storage")
		}
	}

	err = raftMemoryStorage.SetHardState(state)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "failed to set hard state to raft memory storage")
	}

	err = raftMemoryStorage.Append(entries)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "failed to append entries to raft memory storage")
	}

	logger.Info("Loaded previous hardstate and entries to Raft memory storage", zap.Any("hard-state", state), zap.Int("entries-number", len(entries)))

	return &storage{
		logger: logger,

		MemoryStorage: raftMemoryStorage,

		wal: w,

		snapper:   snapper,
		providers: make(map[string]snapshotProvider),
	}, ok, nil
}

func loadSnapshot(logger *zap.Logger, walPath string, snapshotter *snap.Snapshotter) (*raftpb.Snapshot, error) {
	if !wal.Exist(walPath) {
		return &raftpb.Snapshot{}, nil
	}

	entries, err := wal.ValidSnapshotEntries(logger, walPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get valid snapshot entries")
	}

	snapshot, err := snapshotter.LoadNewestAvailable(entries)
	if err != nil {
		if errors.Is(err, snap.ErrNoSnapshot) {
			return &raftpb.Snapshot{}, nil
		}

		return nil, stacktrace.Propagate(err, "failed to load newest available snapshot")
	}

	logger.Info("Loaded snapshot", zap.Uint64("index", snapshot.Metadata.Index), zap.Uint64("term", snapshot.Metadata.Term))
	return snapshot, nil
}

// save saves the given snapshot to the snapshotter and the wal.
func (s *storage) save(snapshot raftpb.Snapshot) error {
	err := s.snapper.SaveSnap(snapshot)
	if err != nil {
		return stacktrace.Propagate(err, "failed to save snapshot")
	}

	err = s.wal.SaveSnapshot(walpb.Snapshot{
		Index:     snapshot.Metadata.Index,
		Term:      snapshot.Metadata.Term,
		ConfState: &snapshot.Metadata.ConfState,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to save snapshot to wal")
	}

	return s.wal.ReleaseLockTo(snapshot.Metadata.Index)
}

func (s *storage) registerSnapshotProvider(name string, provider func() ([]byte, error)) {
	s.providers[name] = provider
}

// getSnapshot calls all registered snapshot providers and combines their data into a single snapshot.
func (s *storage) getSnapshot() ([]byte, error) {
	parts := make(map[string][]byte)
	for name, provider := range s.providers {
		data, err := provider()
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get snapshot data from %q", name)
		}

		parts[name] = data
	}

	return json.Marshal(parts)
}

// snapshotter returns the snapshotter used by the storage.
func (s *storage) snapshotter() *snap.Snapshotter {
	return s.snapper
}

func (s *storage) triggerSnapshot(appliedIndex uint64, confState *raftpb.ConfState) error {
	s.logger.Info("triggering snapshot", zap.Uint64("appliedIndex", appliedIndex))
	data, err := s.getSnapshot()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get snapshot data")
	}

	snap, err := s.CreateSnapshot(appliedIndex, confState, data)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create snapshot from raft memory storage")
	}

	err = s.save(snap)
	if err != nil {
		return stacktrace.Propagate(err, "failed to save snapshot")
	}

	compactIndex := uint64(1)
	if appliedIndex > snapshotCatchUpEntriesN {
		compactIndex = appliedIndex - snapshotCatchUpEntriesN
	}

	err = s.Compact(compactIndex)
	if err != nil && !errors.Is(err, raft.ErrCompacted) {
		return stacktrace.Propagate(err, "failed to compact raft memory storage at index %d", compactIndex)
	}

	return nil
}

func (s *storage) handleReceivedState(snapshot raftpb.Snapshot, hardState raftpb.HardState, entries []raftpb.Entry) error {
	if !raft.IsEmptySnap(snapshot) {
		err := s.save(snapshot)
		if err != nil {
			return stacktrace.Propagate(err, "failed to save snapshot")
		}
	}

	err := s.wal.Save(hardState, entries)
	if err != nil {
		return stacktrace.Propagate(err, "failed to save WAL entries")
	}

	if !raft.IsEmptySnap(snapshot) {
		err := s.ApplySnapshot(snapshot)
		if err != nil {
			return stacktrace.Propagate(err, "failed to apply snapshot")
		}
	}

	if err := s.Append(entries); err != nil {
		return stacktrace.Propagate(err, "failed to append entries to raft storage")
	}

	return nil
}
