package consensus

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
)

func TestNewStorageRepairsTruncatedWAL(t *testing.T) {
	dir := t.TempDir()

	s, nodeID, _, err := newStorage(t.Context(), zap.NewNop(), dir, 0)
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

	s, gotNodeID, _, err := newStorage(t.Context(), zap.NewNop(), dir, 0)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, s.wal.Close()) })

	require.Equal(t, nodeID, gotNodeID)
	lastIndex, err := s.LastIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(n-1), lastIndex)
	require.FileExists(t, last+".broken")
}
