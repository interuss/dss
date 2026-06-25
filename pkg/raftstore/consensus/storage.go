package consensus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"
	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	"go.etcd.io/etcd/server/v3/storage/wal"
	"go.etcd.io/etcd/server/v3/storage/wal/walpb"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
)

// membersFileName is the name of the file, relative to a node's DataDir, that persists the
// current cluster membership (node ID -> peer URL) across restarts.
const membersFileName = "members.json"

// snapshotProvider is a function that returns the snapshot data to be included in the Raft snapshot.
// We use a snapshotProvider from each component (scd, rid and aux).
type snapshotProvider func() ([]byte, error)

// storage persists the Raft log and snapshots and manages the raft in-memory storage.
type storage struct {
	logger *zap.Logger

	*raft.MemoryStorage

	wal *wal.WAL

	snapper  *snap.Snapshotter
	snapshot snapshotProvider

	snapshotCatchUpEntries uint64

	dataDir string
}

// newStorage initializes the storage by loading the latest snapshot and wal entries from the disk
// and applies them to the Raft memory storage.
// It returns the initialized storage, a boolean indicating whether the storage was pre-existent or an error.
func newStorage(ctx context.Context, logger *zap.Logger, dataDir string, nodeID uint64, provider snapshotProvider, snapshotCatchUpEntries uint64) (*storage, bool, error) {
	logger = logging.WithValuesFromContext(ctx, logger)

	// load the latest snapshot
	snapshotPath := path.Join(dataDir, fmt.Sprintf("snapshot_%d", nodeID))
	if !fileutil.Exist(snapshotPath) {
		err := os.MkdirAll(snapshotPath, 0o750)
		if err != nil {
			return nil, false, stacktrace.Propagate(err, "failed to create directory for snapshot storage at: %s", snapshotPath)
		}
	}

	walPath := path.Join(dataDir, fmt.Sprintf("wal_%d", nodeID))

	snapper := snap.New(logger, snapshotPath)

	var (
		w        *wal.WAL
		err      error
		snapshot *raftpb.Snapshot
	)
	ok := wal.Exist(walPath)
	if !ok {
		if err = os.MkdirAll(walPath, 0o750); err != nil {
			return nil, false, stacktrace.Propagate(err, "failed to create directory for wal storage at: %s", walPath)
		}
		w, err = wal.Create(logger, walPath, nil)
		if err != nil {
			return nil, false, stacktrace.Propagate(err, "failed to create wal at: %s", walPath)
		}
		if err = w.Close(); err != nil {
			return nil, false, stacktrace.Propagate(err, "failed to close wal at: %s", walPath)
		}
		snapshot = &raftpb.Snapshot{}
	} else {
		snapshot, err = loadSnapshot(logger, walPath, snapper)
		if err != nil {
			return nil, false, stacktrace.Propagate(err, "failed to load snapshot")
		}
	}

	// open the wal at the given snapshot and get all subsequent entries
	w, err = wal.Open(logger, walPath, walpb.Snapshot{
		Index:     snapshot.Metadata.Index,
		Term:      snapshot.Metadata.Term,
		ConfState: &snapshot.Metadata.ConfState,
	})
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "failed to open wal with index %d, term %d and confstate %v at: %s", snapshot.Metadata.Index, snapshot.Metadata.Term, snapshot.Metadata.ConfState, walPath)
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

		snapper:  snapper,
		snapshot: provider,

		snapshotCatchUpEntries: snapshotCatchUpEntries,

		dataDir: dataDir,
	}, ok, nil
}

func loadSnapshot(logger *zap.Logger, walPath string, snapshotter *snap.Snapshotter) (*raftpb.Snapshot, error) {
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

	// ReleaseLockTo releases the os-level flock locks held on WAL segment files
	// for entries covered by the snapshot and that are no longer needed.
	// We only call this on success. If saving the snapshot failed above, the files
	// are still needed and must remain locked.
	//
	// In etcd's implementation this call marks files eligible for deletion
	// by a background cleanup goroutine which we do not have yet.
	//
	// TODO: add a separate goroutine to cleanup old files
	return s.wal.ReleaseLockTo(snapshot.Metadata.Index)
}

func (s *storage) triggerSnapshot(appliedIndex uint64, confState *raftpb.ConfState, members map[uint64]string, removedIDs []uint64) error {
	s.logger.Info("triggering snapshot", zap.Uint64("appliedIndex", appliedIndex))
	appData, err := s.snapshot()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get snapshot data")
	}

	data, err := json.Marshal(snapshotEnvelope{Members: members, RemovedIDs: removedIDs, AppData: appData})
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal snapshot envelope")
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
	if appliedIndex > s.snapshotCatchUpEntries {
		compactIndex = appliedIndex - s.snapshotCatchUpEntries
	}

	return s.compactLog(compactIndex)
}

func (s *storage) compactLog(compactIndex uint64) error {
	err := s.Compact(compactIndex)
	if errors.Is(err, raft.ErrCompacted) {
		s.logger.Warn("log already compacted", zap.Uint64("compactIndex", compactIndex))
		return nil
	}
	if err != nil {
		return stacktrace.Propagate(err, "failed to compact raft memory storage at index %d", compactIndex)
	}

	s.logger.Info("compacted log", zap.Uint64("compactIndex", compactIndex))
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

// saveMembers persists the current node ID -> peer URL table to dataDir/members.json
// so that it can be loaded on restart.
func (s *storage) saveMembers(members map[uint64]string) error {
	data, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal member list")
	}

	tmpPath := path.Join(s.dataDir, membersFileName+".tmp")
	if err := os.WriteFile(tmpPath, data, 0o640); err != nil {
		return stacktrace.Propagate(err, "failed to write member list to %s", tmpPath)
	}

	finalPath := path.Join(s.dataDir, membersFileName)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return stacktrace.Propagate(err, "failed to rename member list into place at %s", finalPath)
	}

	return nil
}

// loadMembers loads the persisted node ID -> peer URL table from dataDir/members.json, if it exists.
func loadMembers(dataDir string) (map[uint64]string, error) {
	membersPath := path.Join(dataDir, membersFileName)
	if !fileutil.Exist(membersPath) {
		return nil, nil
	}

	data, err := os.ReadFile(membersPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to read member list from %s", membersPath)
	}

	members := make(map[uint64]string)
	if err := json.Unmarshal(data, &members); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal member list from %s", membersPath)
	}

	return members, nil
}

func membersToPeerMap(members map[uint64]string) (map[uint64]*url.URL, error) {
	peers := make(map[uint64]*url.URL, len(members))
	for id, rawURL := range members {
		peerURL, err := url.Parse(rawURL)
		if err != nil {
			return nil, stacktrace.Propagate(err, "invalid URL %s for node %d", rawURL, id)
		}
		peers[id] = peerURL
	}
	return peers, nil
}
