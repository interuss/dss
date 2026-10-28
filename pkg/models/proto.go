package models

import (
	"time"
	"uuid"

	"github.com/golang/geo/s2"
	modelspb "github.com/interuss/dss/pkg/pb"
	"github.com/interuss/stacktrace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func IDToProto(id ID) ([]byte, error) {
	parsed, err := uuid.Parse(id.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid ID %q", id)
	}
	return parsed[:], nil
}

func IDFromProto(b []byte) (ID, error) {
	if len(b) != len(uuid.UUID{}) {
		return "", stacktrace.NewError("invalid ID length %d", len(b))
	}
	return ID(uuid.UUID(b).String()), nil
}

func VersionToProto(v *Version) string {
	if v == nil {
		return ""
	}
	return v.String()
}

func VersionFromProto(s string) (*Version, error) {
	if s == "" {
		return nil, nil
	}
	return VersionFromString(s)
}

func TimeToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func TimeFromProto(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func CellUnionToProto(cells s2.CellUnion) []uint64 {
	ids := make([]uint64, len(cells))
	for i, cell := range cells {
		ids[i] = uint64(cell)
	}
	return ids
}

func CellUnionFromProto(ids []uint64) s2.CellUnion {
	cells := make(s2.CellUnion, len(ids))
	for i, id := range ids {
		cells[i] = s2.CellID(id)
	}
	return cells
}

func CellsVolume4DToProto(v *CellsVolume4D) *modelspb.CellsVolume4D {
	if v == nil {
		return nil
	}
	return &modelspb.CellsVolume4D{
		Cells:      CellUnionToProto(v.Cells),
		StartTime:  TimeToProto(v.StartTime),
		EndTime:    TimeToProto(v.EndTime),
		AltitudeLo: v.AltitudeLo,
		AltitudeHi: v.AltitudeHi,
	}
}

func CellsVolume4DFromProto(pb *modelspb.CellsVolume4D) *CellsVolume4D {
	if pb == nil {
		return nil
	}
	return &CellsVolume4D{
		Cells:      CellUnionFromProto(pb.GetCells()),
		StartTime:  TimeFromProto(pb.GetStartTime()),
		EndTime:    TimeFromProto(pb.GetEndTime()),
		AltitudeLo: pb.AltitudeLo,
		AltitudeHi: pb.AltitudeHi,
	}
}
