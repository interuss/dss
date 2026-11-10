package consensus

import (
	"errors"
	"testing"

	proposalpb "github.com/interuss/dss/pkg/pb/raftpb"
	ridsnapshotpb "github.com/interuss/dss/pkg/pb/ridpb/snapshotpb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func ridAppSnapshot() *proposalpb.Snapshot {
	return &proposalpb.Snapshot{AppSnapshot: &proposalpb.Snapshot_RidSnapshot{RidSnapshot: &ridsnapshotpb.Snapshot{
		Isas: map[string]*ridsnapshotpb.ISA{"00000000-0000-4000-8000-000000000001": {Url: "https://uss.example.com/isa"}},
	}}}
}

func consensusWithSnapshot(members, removed memberMap, provider func() (*proposalpb.Snapshot, error)) *Consensus {
	return &Consensus{
		membership:     membership{members: members, removed: removed},
		getAppSnapshot: provider,
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		members memberMap
		removed memberMap
	}{
		{
			name:    "members and removed",
			members: memberMap{1: {URL: "https://node1:9011"}, 2: {URL: "https://node2:9011"}},
			removed: memberMap{3: {URL: "https://node3:9011"}},
		},
		{
			name:    "no removed",
			members: memberMap{1: {URL: "https://node1:9011"}},
			removed: memberMap{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := consensusWithSnapshot(tt.members, tt.removed, func() (*proposalpb.Snapshot, error) { return ridAppSnapshot(), nil })

			data, err := c.encodeSnapshot()
			require.NoError(t, err)
			got, err := decodeSnapshot(data)
			require.NoError(t, err)

			require.Equal(t, tt.members, got.Members)
			require.Equal(t, tt.removed, got.Removed)
			require.True(t, proto.Equal(ridAppSnapshot().GetRidSnapshot(), got.AppSnapshot.GetRidSnapshot()))
		})
	}
}

func TestEncodeSnapshotErrors(t *testing.T) {
	members := memberMap{1: {URL: "https://node1:9011"}}

	_, err := consensusWithSnapshot(members, memberMap{}, func() (*proposalpb.Snapshot, error) {
		return nil, errors.New("snapshot failed")
	}).encodeSnapshot()
	require.Error(t, err)

	_, err = consensusWithSnapshot(members, memberMap{}, func() (*proposalpb.Snapshot, error) {
		// No AppSnapshot
		return &proposalpb.Snapshot{}, nil
	}).encodeSnapshot()
	require.Error(t, err)
}

func TestDecodeSnapshotErrors(t *testing.T) {
	_, err := decodeSnapshot([]byte("invalid data"))
	require.Error(t, err)

	data, err := proto.Marshal(&proposalpb.Snapshot{Members: map[uint64]*proposalpb.Member{1: {PeerUrl: "https://node1:9011"}}})
	require.NoError(t, err)
	_, err = decodeSnapshot(data)
	require.Error(t, err)
}
