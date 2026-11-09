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

func TestConstraintProtoRoundTrip(t *testing.T) {
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

	for name, constraint := range map[string]*Constraint{
		"all fields": {
			ID:            dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a"),
			Manager:       dssmodels.Manager("uss1"),
			USSBaseURL:    "https://uss.example.com",
			CellsVolume4D: volume,
		},
		"empty fields": {
			ID:            dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a"),
			CellsVolume4D: volume,
		},
	} {
		t.Run(name, func(t *testing.T) {
			pb, err := constraint.ToProto()
			require.NoError(t, err)
			buf, err := proto.Marshal(pb)
			require.NoError(t, err)
			var decodedPb scdpb.Constraint
			require.NoError(t, proto.Unmarshal(buf, &decodedPb))

			var decoded Constraint
			require.NoError(t, decoded.FromProto(&decodedPb))
			assert.Equal(t, constraint, &decoded)
		})
	}
}

func TestConstraintProtoErrors(t *testing.T) {
	var missing *Constraint
	_, err := missing.ToProto()
	require.Error(t, err)

	_, err = (&Constraint{ID: dssmodels.ID("not-a-uuid"), CellsVolume4D: &dssmodels.CellsVolume4D{}}).ToProto()
	require.Error(t, err)

	_, err = (&Constraint{ID: dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a")}).ToProto()
	require.Error(t, err)

	var constraint Constraint
	require.Error(t, constraint.FromProto(nil))
	require.Error(t, constraint.FromProto(&scdpb.Constraint{Id: []byte{1, 2, 3}, Volume: (&dssmodels.CellsVolume4D{}).ToProto()}))
	id, err := dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a").ToProto()
	require.NoError(t, err)
	require.Error(t, constraint.FromProto(&scdpb.Constraint{Id: id}))
}
