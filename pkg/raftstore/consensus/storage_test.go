package consensus

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	params "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/server/v3/storage/wal"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
)

func TestNewStorageRepairsTruncatedWAL(t *testing.T) {
	dir := t.TempDir()

	s, nodeID, _, err := newStorage(t.Context(), zap.NewNop(), dir, 0, params.PurgeParameters{})
	require.NoError(t, err)

	const n = 10
	for i := uint64(1); i <= n; i++ {
		require.NoError(t, s.wal.Save(&raftpb.HardState{}, []*raftpb.Entry{{Index: new(i), Data: []byte("payload")}}))
	}
	require.NoError(t, s.wal.Close())

	segments, err := filepath.Glob(filepath.Join(dir, "wal", "*.wal"))
	require.NoError(t, err)
	require.NotEmpty(t, segments)
	last := segments[len(segments)-1]
	data, err := os.ReadFile(last)
	require.NoError(t, err)
	require.NoError(t, os.Truncate(last, int64(len(bytes.TrimRight(data, "\x00"))-4)))

	s, gotNodeID, _, err := newStorage(t.Context(), zap.NewNop(), dir, 0, params.PurgeParameters{})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, s.wal.Close()) })

	require.Equal(t, nodeID, gotNodeID)
	lastIndex, err := s.LastIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(n-1), lastIndex)
	require.FileExists(t, last+".broken")
}

const purgeInterval = 10 * time.Millisecond

// newPurgedStorage returns a storage holding n entries spread over multiple WAL segments.
func newPurgedStorage(t *testing.T, dir string, n uint64, purge params.PurgeParameters) *storage {
	segmentSize := wal.SegmentSizeBytes
	wal.SegmentSizeBytes = 4 * 1024
	t.Cleanup(func() { wal.SegmentSizeBytes = segmentSize })

	s, _, _, err := newStorage(t.Context(), zap.NewNop(), dir, 0, purge)
	require.NoError(t, err)
	for i := uint64(1); i <= n; i++ {
		entries := []*raftpb.Entry{{Index: new(i), Term: new(uint64(1)), Data: bytes.Repeat([]byte("x"), 1024)}}
		require.NoError(t, s.handleReceivedState(&raftpb.Snapshot{}, &raftpb.HardState{Term: new(uint64(1)), Commit: new(i)}, entries))
	}
	return s
}

func countFiles(t *testing.T, dir, pattern string) int {
	files, err := filepath.Glob(filepath.Join(dir, pattern))
	require.NoError(t, err)
	return len(files)
}

const (
	maxWALs      = 2
	maxSnapshots = 2

	walFilesPattern  = "wal/*.wal"
	snapFilesPattern = "snap/*.snap"
)

func TestPurgeDeletesReleasedWALSegmentsAndOldSnapshots(t *testing.T) {
	dir := t.TempDir()
	s := newPurgedStorage(t, dir, 40, params.PurgeParameters{MaxWALs: maxWALs, MaxSnapshots: maxSnapshots, Interval: purgeInterval})
	t.Cleanup(func() { require.NoError(t, s.close()) })

	segments := countFiles(t, dir, walFilesPattern)
	require.Greater(t, segments, maxWALs)
	time.Sleep(5 * purgeInterval)

	// check that the purge has not deleted any segments that are still locked by the WAL
	require.Equal(t, segments, countFiles(t, dir, walFilesPattern), "segments still locked by the WAL must be kept")

	for _, index := range []uint64{10, 20, 30, 40} {
		require.NoError(t, s.triggerSnapshot(index, &raftpb.ConfState{}, nil))
	}

	require.Eventually(t, func() bool {
		return countFiles(t, dir, walFilesPattern) <= maxWALs && countFiles(t, dir, snapFilesPattern) == maxSnapshots
	}, time.Second, purgeInterval)
}

func TestStorageCloseStopsPurge(t *testing.T) {
	dir := t.TempDir()
	s := newPurgedStorage(t, dir, 40, params.PurgeParameters{MaxWALs: maxWALs, Interval: purgeInterval})

	segments := countFiles(t, dir, walFilesPattern)
	require.Greater(t, segments, maxWALs)
	require.NoError(t, s.close())

	// closing the WAL releases all the flocks, so a purge still running would delete the segments
	time.Sleep(5 * purgeInterval)
	require.Equal(t, segments, countFiles(t, dir, walFilesPattern))
}
