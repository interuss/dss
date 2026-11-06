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
		"uuid":  ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a"),
		"empty": ID(""),
	} {
		t.Run(name, func(t *testing.T) {
			b, err := IDToProto(id)
			require.NoError(t, err)
			decoded, err := IDFromProto(b)
			require.NoError(t, err)
			require.Equal(t, id, decoded)
		})
	}
}

func TestIDProtoErrors(t *testing.T) {
	_, err := IDToProto(ID("not-a-uuid"))
	require.Error(t, err)

	_, err = IDFromProto([]byte{1, 2, 3})
	require.Error(t, err)
}

func TestVersionProtoRoundTrip(t *testing.T) {
	version := VersionFromTime(time.Date(2026, 10, 28, 9, 30, 0, 123456789, time.UTC))
	decoded, err := VersionFromProto(VersionToProto(version))
	require.NoError(t, err)
	require.Equal(t, version.String(), decoded.String())
}

func TestVersionProtoEmpty(t *testing.T) {
	require.Empty(t, VersionToProto(nil))

	version, err := VersionFromProto("")
	require.NoError(t, err)
	require.Nil(t, version)
}

func TestVersionProtoErrors(t *testing.T) {
	_, err := VersionFromProto("not a version")
	require.Error(t, err)
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
		"empty":                  {},
	} {
		t.Run(name, func(t *testing.T) {
			buf, err := proto.Marshal(CellsVolume4DToProto(volume))
			require.NoError(t, err)
			var pb modelspb.CellsVolume4D
			require.NoError(t, proto.Unmarshal(buf, &pb))

			require.Equal(t, volume, CellsVolume4DFromProto(&pb))
		})
	}

	require.Nil(t, CellsVolume4DToProto(nil))
	require.Nil(t, CellsVolume4DFromProto(nil))
}
