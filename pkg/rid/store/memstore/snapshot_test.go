package memstore

import (
	"bytes"
	"context"
	"encoding/gob"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/interuss/dss/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestSnapshotRoundTrip(t *testing.T) {
	ctx := context.Background()
	src := setUpStore(t)
	_, err := src.InsertISA(ctx, serviceArea)
	require.NoError(t, err)
	_, err = src.InsertSubscription(ctx, subscriptionsPool[0].input)
	require.NoError(t, err)

	data, err := src.GetSnapshot()
	require.NoError(t, err)

	dst := setUpStore(t)
	require.NoError(t, dst.RestoreFromSnapshot(data))

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

func TestRestoreFromSnapshotInvalidData(t *testing.T) {
	require.Error(t, setUpStore(t).RestoreFromSnapshot([]byte("random value that is definitely not valid")))
}

func TestRestoreFromSnapshotVersionMismatch(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(snapshotEnvelope{Version: snapshotVersion + 1}))
	require.Error(t, setUpStore(t).RestoreFromSnapshot(buf.Bytes()))
}
