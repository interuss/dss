package memstore

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
)

func TestSnapshotRoundTrip(t *testing.T) {
	ctx := writeCtx()
	src := setUpStore(t)
	_, err := src.UpsertConstraint(ctx, sampleConstraint())
	require.NoError(t, err)
	_, err = src.UpsertSubscription(ctx, sampleSubscription())
	require.NoError(t, err)
	_, err = src.UpsertOperationalIntent(ctx, sampleOperationalIntent())
	require.NoError(t, err)
	_, err = src.UpsertUssAvailability(ctx, sampleAvailability())
	require.NoError(t, err)

	data, err := src.GetSnapshot()
	require.NoError(t, err)

	dst := setUpStore(t)
	require.NoError(t, dst.RestoreFromSnapshot(data))

	opt := cmpopts.EquateApproxTime(0)

	wantCon, err := src.GetConstraint(ctx, constraintId)
	require.NoError(t, err)
	gotCon, err := dst.GetConstraint(ctx, constraintId)
	require.NoError(t, err)
	if diff := cmp.Diff(wantCon, gotCon, opt); diff != "" {
		t.Errorf("Constraint mismatch (-want +got):\n%s", diff)
	}

	wantSub, err := src.GetSubscription(ctx, subscriptionId)
	require.NoError(t, err)
	gotSub, err := dst.GetSubscription(ctx, subscriptionId)
	require.NoError(t, err)
	if diff := cmp.Diff(wantSub, gotSub, opt); diff != "" {
		t.Errorf("Subscription mismatch (-want +got):\n%s", diff)
	}

	wantOI, err := src.GetOperationalIntent(ctx, operationalIntentId)
	require.NoError(t, err)
	gotOI, err := dst.GetOperationalIntent(ctx, operationalIntentId)
	require.NoError(t, err)
	if diff := cmp.Diff(wantOI, gotOI, opt); diff != "" {
		t.Errorf("OperationalIntent mismatch (-want +got):\n%s", diff)
	}

	wantAvail, err := src.GetUssAvailability(ctx, manager)
	require.NoError(t, err)
	gotAvail, err := dst.GetUssAvailability(ctx, manager)
	require.NoError(t, err)
	if diff := cmp.Diff(wantAvail, gotAvail, opt); diff != "" {
		t.Errorf("UssAvailability mismatch (-want +got):\n%s", diff)
	}
}

func TestRestoreFromSnapshotReplacesState(t *testing.T) {
	ctx := writeCtx()
	src := setUpStore(t)
	_, err := src.UpsertConstraint(ctx, sampleConstraint())
	require.NoError(t, err)
	data, err := src.GetSnapshot()
	require.NoError(t, err)

	dst := setUpStore(t)
	other := sampleConstraint()
	other.ID = "00000185-e36d-40be-8d38-beca6ca39999"
	_, err = dst.UpsertConstraint(ctx, other)
	require.NoError(t, err)
	require.NoError(t, dst.RestoreFromSnapshot(data))

	count, err := dst.CountConstraints(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
	got, err := dst.GetConstraint(ctx, constraintId)
	require.NoError(t, err)
	require.NotNil(t, got)
	_, err = dst.GetConstraint(ctx, other.ID)
	require.Error(t, err)
}

func TestRestoreFromSnapshotInvalidData(t *testing.T) {
	require.Error(t, setUpStore(t).RestoreFromSnapshot([]byte("random value that is definitely not valid")))
}

func TestRestoreFromSnapshotVersionMismatch(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(snapshotEnvelope{Version: snapshotVersion + 1}))
	require.Error(t, setUpStore(t).RestoreFromSnapshot(buf.Bytes()))
}
