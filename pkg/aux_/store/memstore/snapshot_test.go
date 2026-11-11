package memstore

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	"github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/raftpb"
	ridsnapshotpb "github.com/interuss/dss/pkg/pb/ridpb/snapshotpb"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSnapshotRoundTrip(t *testing.T) {
	ctx := context.Background()
	ctx = timestamp.NewContext(ctx, fakeClock.Now())
	src := newRepo()
	require.NoError(t, src.SaveOwnMetadata(ctx, "dss-1", "https://example.com"))
	ts := time.Now().UTC()
	require.NoError(t, src.RecordHeartbeat(ctx, auxmodels.Heartbeat{Locality: "dss-1", Source: "source-1", Timestamp: &ts, Reporter: "uss-1"}))

	data, err := src.GetSnapshot()
	require.NoError(t, err)

	dst := newRepo()
	require.NoError(t, dst.RestoreFromSnapshot(encodeDecode(t, data)))

	want, err := src.GetDSSMetadata(ctx)
	require.NoError(t, err)
	got, err := dst.GetDSSMetadata(ctx)
	require.NoError(t, err)
	if diff := cmp.Diff(want, got, cmpopts.EquateApproxTime(0), cmp.AllowUnexported(models.Version{})); diff != "" {
		t.Errorf("Store mismatch (-want +got):\n%s", diff)
	}
}

func TestRestoreFromSnapshotReplacesState(t *testing.T) {
	ctx := context.Background()
	ctx = timestamp.NewContext(ctx, fakeClock.Now())
	src := newRepo()
	require.NoError(t, src.SaveOwnMetadata(ctx, "dss-1", "https://example.com"))
	data, err := src.GetSnapshot()
	require.NoError(t, err)

	dst := newRepo()
	require.NoError(t, dst.SaveOwnMetadata(ctx, "dss-2", "https://other.example.com"))
	require.NoError(t, dst.RestoreFromSnapshot(data))

	md, err := dst.GetDSSMetadata(ctx)
	require.NoError(t, err)
	require.Len(t, md, 1)
	require.Equal(t, "dss-1", md[0].Locality)
}

func TestRestoreFromSnapshotWrongType(t *testing.T) {
	require.Error(t, newRepo().RestoreFromSnapshot(&raftpb.Snapshot{}))
	require.Error(t, newRepo().RestoreFromSnapshot(&raftpb.Snapshot{AppSnapshot: &raftpb.Snapshot_RidSnapshot{RidSnapshot: &ridsnapshotpb.Snapshot{}}}))
}

func encodeDecode(t *testing.T, snap *raftpb.Snapshot) *raftpb.Snapshot {
	t.Helper()
	data, err := proto.Marshal(snap)
	require.NoError(t, err)
	var decoded raftpb.Snapshot
	require.NoError(t, proto.Unmarshal(data, &decoded))
	return &decoded
}
