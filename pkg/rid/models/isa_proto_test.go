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
	version := dssmodels.VersionFromTime(start)
	id := dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a")

	for name, isa := range map[string]*IdentificationServiceArea{
		"all fields": {
			ID:            id,
			URL:           "https://uss.example.com/isa",
			Owner:         dssmodels.Owner("uss1"),
			Version:       version,
			Writer:        "writer",
			CellsVolume4D: volume,
		},
		"no version": {
			ID:            id,
			Owner:         dssmodels.Owner("uss1"),
			CellsVolume4D: volume,
		},
	} {
		t.Run(name, func(t *testing.T) {
			pb, err := ISAToProto(isa)
			require.NoError(t, err)
			buf, err := proto.Marshal(pb)
			require.NoError(t, err)
			var decodedPb ridpb.IdentificationServiceArea
			require.NoError(t, proto.Unmarshal(buf, &decodedPb))

			decoded, err := ISAFromProto(&decodedPb, dssmodels.VersionToProto(isa.Version))
			require.NoError(t, err)
			assert.Equal(t, isa.ID, decoded.ID)
			assert.Equal(t, isa.URL, decoded.URL)
			assert.Equal(t, isa.Owner, decoded.Owner)
			assert.Equal(t, isa.Writer, decoded.Writer)
			assert.Equal(t, isa.CellsVolume4D, decoded.CellsVolume4D)
			assert.Equal(t, dssmodels.VersionToProto(isa.Version), dssmodels.VersionToProto(decoded.Version))
		})
	}

	pb, err := ISAToProto(nil)
	require.NoError(t, err)
	assert.Nil(t, pb)
	isa, err := ISAFromProto(nil, "")
	require.NoError(t, err)
	assert.Nil(t, isa)
}

func TestISAProtoInvalid(t *testing.T) {
	_, err := ISAToProto(&IdentificationServiceArea{ID: dssmodels.ID("not-a-uuid")})
	require.Error(t, err)

	_, err = ISAFromProto(&ridpb.IdentificationServiceArea{Id: []byte{1, 2, 3}}, "")
	require.Error(t, err)
}
