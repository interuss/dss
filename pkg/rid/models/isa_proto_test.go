package models

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/ridpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestISAProtoRoundTrip(t *testing.T) {
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
	id := dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a")

	for name, isa := range map[string]*IdentificationServiceArea{
		"all fields": {
			ID:            id,
			URL:           "https://uss.example.com/isa",
			Owner:         dssmodels.Owner("uss1"),
			Writer:        "writer",
			CellsVolume4D: volume,
		},
		"empty fields": {
			ID:            id,
			Owner:         dssmodels.Owner("uss1"),
			CellsVolume4D: volume,
		},
	} {
		t.Run(name, func(t *testing.T) {
			pb, err := isa.ToProto()
			require.NoError(t, err)
			buf, err := proto.Marshal(pb)
			require.NoError(t, err)
			var decodedPb ridpb.IdentificationServiceArea
			require.NoError(t, proto.Unmarshal(buf, &decodedPb))

			var decoded IdentificationServiceArea
			require.NoError(t, decoded.FromProto(&decodedPb))
			assert.Equal(t, isa, &decoded)
		})
	}
}

func TestISAProtoErrors(t *testing.T) {
	var missing *IdentificationServiceArea
	_, err := missing.ToProto()
	require.Error(t, err)

	_, err = (&IdentificationServiceArea{ID: dssmodels.ID("not-a-uuid"), CellsVolume4D: &dssmodels.CellsVolume4D{}}).ToProto()
	require.Error(t, err)

	_, err = (&IdentificationServiceArea{ID: dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a")}).ToProto()
	require.Error(t, err)

	var isa IdentificationServiceArea
	require.Error(t, isa.FromProto(nil))
	require.Error(t, isa.FromProto(&ridpb.IdentificationServiceArea{Id: []byte{1, 2, 3}, Volume: (&dssmodels.CellsVolume4D{}).ToProto()}))
	id, err := dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a").ToProto()
	require.NoError(t, err)
	require.Error(t, isa.FromProto(&ridpb.IdentificationServiceArea{Id: id}))
}
