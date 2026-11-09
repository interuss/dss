package models

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestOperationalIntentProtoRoundTrip(t *testing.T) {
	start := time.Date(2026, 10, 28, 9, 30, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	lo, hi := float32(10), float32(120)
	volume := &dssmodels.CellsVolume4D{
		Cells:      s2.CellUnion{s2.CellIDFromToken("89c25"), s2.CellIDFromToken("89c27")},
		StartTime:  &start,
		EndTime:    &end,
		AltitudeLo: &lo,
		AltitudeHi: &hi,
	}
	subID := dssmodels.ID("5d0f0c4e-1f7a-4b0e-9d3c-2a8b6e4f1c7d")

	for name, op := range map[string]*OperationalIntent{
		"all fields": {
			ID:             dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a"),
			Manager:        dssmodels.Manager("uss1"),
			Version:        3,
			State:          OperationalIntentStateContingent,
			OVN:            OVN("custom-ovn"),
			PastOVNs:       []OVN{"0192e4b8-7a00-7000-8000-000000000000_0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a", "other"},
			USSBaseURL:     "https://uss.example.com",
			SubscriptionID: &subID,
			CellsVolume4D:  volume,
		},
		"empty fields": {
			ID:            dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a"),
			CellsVolume4D: volume,
		},
	} {
		t.Run(name, func(t *testing.T) {
			pb, err := op.ToProto()
			require.NoError(t, err)
			buf, err := proto.Marshal(pb)
			require.NoError(t, err)
			var decodedPb scdpb.UpsertOperationalIntentRepoCall_OperationalIntent
			require.NoError(t, proto.Unmarshal(buf, &decodedPb))

			var decoded OperationalIntent
			require.NoError(t, decoded.FromProto(&decodedPb))
			assert.Equal(t, op, &decoded)
		})
	}
}

func TestOperationalIntentProtoErrors(t *testing.T) {
	validID := dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a")
	badID := dssmodels.ID("not-a-uuid")

	var missing *OperationalIntent
	_, err := missing.ToProto()
	require.Error(t, err)

	_, err = (&OperationalIntent{ID: badID, CellsVolume4D: &dssmodels.CellsVolume4D{}}).ToProto()
	require.Error(t, err)

	_, err = (&OperationalIntent{ID: validID}).ToProto()
	require.Error(t, err)

	_, err = (&OperationalIntent{ID: validID, SubscriptionID: &badID, CellsVolume4D: &dssmodels.CellsVolume4D{}}).ToProto()
	require.Error(t, err)

	_, err = (&OperationalIntent{ID: validID, State: "Bogus", CellsVolume4D: &dssmodels.CellsVolume4D{}}).ToProto()
	require.Error(t, err)

	id, err := validID.ToProto()
	require.NoError(t, err)
	volume := (&dssmodels.CellsVolume4D{}).ToProto()

	var op OperationalIntent
	require.Error(t, op.FromProto(nil))
	require.Error(t, op.FromProto(&scdpb.UpsertOperationalIntentRepoCall_OperationalIntent{Id: []byte{1, 2, 3}, Volume: volume}))
	require.Error(t, op.FromProto(&scdpb.UpsertOperationalIntentRepoCall_OperationalIntent{Id: id}))
	require.Error(t, op.FromProto(&scdpb.UpsertOperationalIntentRepoCall_OperationalIntent{Id: id, Volume: volume, SubscriptionId: []byte{1}}))
	require.Error(t, op.FromProto(&scdpb.UpsertOperationalIntentRepoCall_OperationalIntent{Id: id, Volume: volume, State: 42}))
}
