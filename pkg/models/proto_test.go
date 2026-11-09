package models

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	modelspb "github.com/interuss/dss/pkg/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestIDProtoRoundTrip(t *testing.T) {
	for name, id := range map[string]ID{
		"uuid": ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a"),
	} {
		t.Run(name, func(t *testing.T) {
			b, err := id.ToProto()
			require.NoError(t, err)
			var decoded ID
			require.NoError(t, decoded.FromProto(b))
			require.Equal(t, id, decoded)
		})
	}
}

func TestIDProtoErrors(t *testing.T) {
	_, err := ID("not-a-uuid").ToProto()
	require.Error(t, err)

	_, err = ID("").ToProto()
	require.Error(t, err)

	var id ID
	require.Error(t, id.FromProto([]byte{1, 2, 3}))
	require.Error(t, id.FromProto(nil))
}

func TestVersionProtoRoundTrip(t *testing.T) {
	version := VersionFromTime(time.Date(2026, 10, 28, 9, 30, 0, 123456789, time.UTC))
	var decoded Version
	require.NoError(t, decoded.FromProto(version.ToProto()))
	require.Equal(t, version.String(), decoded.String())
}

func TestVersionProtoEmpty(t *testing.T) {
	var version *Version
	require.Empty(t, version.ToProto())

	require.Error(t, new(Version).FromProto(""))
}

func TestVersionProtoErrors(t *testing.T) {
	require.Error(t, new(Version).FromProto("not a version"))
}

func TestCellsVolume4DProtoRoundTrip(t *testing.T) {
	start := time.Date(2026, 10, 28, 9, 30, 0, 123456789, time.UTC)
	end := start.Add(time.Hour)
	lo, hi := float32(10.5), float32(120)

	for name, volume := range map[string]*CellsVolume4D{
		"all fields": {
			Cells:      s2.CellUnion{s2.CellIDFromToken("89c25"), s2.CellIDFromToken("89c27"), s2.CellIDFromToken("89c3")},
			StartTime:  &start,
			EndTime:    &end,
			AltitudeLo: &lo,
			AltitudeHi: &hi,
		},
		"no times nor altitudes": {Cells: s2.CellUnion{s2.CellIDFromToken("89c25")}},
		"empty":                  {Cells: s2.CellUnion{}},
	} {
		t.Run(name, func(t *testing.T) {
			buf, err := proto.Marshal(volume.ToProto())
			require.NoError(t, err)
			var pb modelspb.CellsVolume4D
			require.NoError(t, proto.Unmarshal(buf, &pb))

			var decoded CellsVolume4D
			decoded.FromProto(&pb)
			require.Equal(t, volume, &decoded)
		})
	}

	var volume *CellsVolume4D
	require.Nil(t, volume.ToProto())
	var decoded CellsVolume4D
	decoded.FromProto(nil)
	require.Equal(t, CellsVolume4D{}, decoded)
}
