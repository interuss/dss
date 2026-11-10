package memstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/raftpb"
	scdsnapshotpb "github.com/interuss/dss/pkg/pb/scdpb/snapshotpb"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSnapshotRoundTrip(t *testing.T) {
	ctx := context.Background()
	ctx = timestamp.NewContext(ctx, fakeClock.Now())
	src := setUpStore(t)
	_, err := src.InsertISA(ctx, serviceArea)
	require.NoError(t, err)
	_, err = src.InsertSubscription(ctx, subscriptionsPool[0].input)
	require.NoError(t, err)

	data, err := src.GetSnapshot()
	require.NoError(t, err)

	dst := setUpStore(t)
	require.NoError(t, dst.RestoreFromSnapshot(encodeDecode(t, data)))

	wantISA, err := src.GetISA(ctx, serviceArea.ID, false)
	require.NoError(t, err)
	gotISA, err := dst.GetISA(ctx, serviceArea.ID, false)
	require.NoError(t, err)

	if diff := cmp.Diff(wantISA, gotISA, cmpopts.EquateApproxTime(0), cmp.AllowUnexported(models.Version{})); diff != "" {
		t.Errorf("IdentificationServiceArea mismatch (-want +got):\n%s", diff)
	}

	wantSub, err := src.GetSubscription(ctx, subscriptionsPool[0].input.ID)
	require.NoError(t, err)
	gotSub, err := dst.GetSubscription(ctx, subscriptionsPool[0].input.ID)
	require.NoError(t, err)

	if diff := cmp.Diff(wantSub, gotSub, cmpopts.EquateApproxTime(0), cmp.AllowUnexported(models.Version{})); diff != "" {
		t.Errorf("Subscription mismatch (-want +got):\n%s", diff)
	}
}

func TestRestoreFromSnapshotReplacesState(t *testing.T) {
	ctx := context.Background()
	ctx = timestamp.NewContext(ctx, fakeClock.Now())
	src := setUpStore(t)
	_, err := src.InsertISA(ctx, serviceArea)
	require.NoError(t, err)
	data, err := src.GetSnapshot()
	require.NoError(t, err)

	dst := setUpStore(t)
	other := *serviceArea
	other.ID = "00000000-0000-4000-8000-000000000002"
	_, err = dst.InsertISA(ctx, &other)
	require.NoError(t, err)
	require.NoError(t, dst.RestoreFromSnapshot(data))

	count, err := dst.CountISAs(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
	got, err := dst.GetISA(ctx, serviceArea.ID, false)
	require.NoError(t, err)
	require.NotNil(t, got)
	gone, err := dst.GetISA(ctx, other.ID, false)
	require.NoError(t, err)
	require.Nil(t, gone)
}

func TestRestoreFromSnapshotWrongType(t *testing.T) {
	require.Error(t, setUpStore(t).RestoreFromSnapshot(&raftpb.Snapshot{}))
	require.Error(t, setUpStore(t).RestoreFromSnapshot(&raftpb.Snapshot{AppSnapshot: &raftpb.Snapshot_ScdSnapshot{ScdSnapshot: &scdsnapshotpb.Snapshot{}}}))
}

func TestSnapshotPreservesAltitudes(t *testing.T) {
	zero := float32(0)
	for name, alts := range map[string][2]*float32{
		"nil":     {nil, nil},
		"zero lo": {&zero, nil},
		"zero hi": {nil, &zero},
		"zero":    {&zero, &zero},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := timestamp.NewContext(context.Background(), fakeClock.Now())
			src := setUpStore(t)
			isa := *serviceArea
			isa.CellsVolume4D = isa.Clone()
			isa.AltitudeLo, isa.AltitudeHi = alts[0], alts[1]
			_, err := src.InsertISA(ctx, &isa)
			require.NoError(t, err)
			sub := *subscriptionsPool[0].input
			sub.CellsVolume4D = sub.Clone()
			sub.AltitudeLo, sub.AltitudeHi = alts[0], alts[1]
			_, err = src.InsertSubscription(ctx, &sub)
			require.NoError(t, err)

			snap, err := src.GetSnapshot()
			require.NoError(t, err)
			dst := setUpStore(t)
			require.NoError(t, dst.RestoreFromSnapshot(encodeDecode(t, snap)))

			gotISA, err := dst.GetISA(ctx, isa.ID, false)
			require.NoError(t, err)
			require.Equal(t, alts[0], gotISA.AltitudeLo)
			require.Equal(t, alts[1], gotISA.AltitudeHi)
			gotSub, err := dst.GetSubscription(ctx, sub.ID)
			require.NoError(t, err)
			require.Equal(t, alts[0], gotSub.AltitudeLo)
			require.Equal(t, alts[1], gotSub.AltitudeHi)
		})
	}
}

func encodeDecode(t *testing.T, snap *raftpb.Snapshot) *raftpb.Snapshot {
	t.Helper()
	data, err := proto.Marshal(snap)
	require.NoError(t, err)
	var decoded raftpb.Snapshot
	require.NoError(t, proto.Unmarshal(data, &decoded))
	return &decoded
}
