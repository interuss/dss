package models

import (
	"time"
	"uuid"

	"github.com/golang/geo/s2"
	modelspb "github.com/interuss/dss/pkg/pb"
	"github.com/interuss/stacktrace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (id ID) ToProto() ([]byte, error) {
	parsed, err := uuid.Parse(id.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid ID %q", id)
	}
	return parsed[:], nil
}

func (id *ID) FromProto(b []byte) error {
	if len(b) != len(uuid.UUID{}) {
		return stacktrace.NewError("invalid ID length %d", len(b))
	}
	*id = ID(uuid.UUID(b).String())
	return nil
}

func (v *Version) ToProto() string {
	return v.String()
}

func (v *Version) FromProto(s string) error {
	parsed, err := VersionFromString(s)
	if err != nil {
		return stacktrace.Propagate(err, "invalid version %q", s)
	}
	*v = *parsed
	return nil
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

func (v *CellsVolume4D) ToProto() *modelspb.CellsVolume4D {
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

func (v *CellsVolume4D) FromProto(pb *modelspb.CellsVolume4D) {
	if pb == nil {
		*v = CellsVolume4D{}
		return
	}
	*v = CellsVolume4D{
		Cells:      CellUnionFromProto(pb.GetCells()),
		StartTime:  TimeFromProto(pb.GetStartTime()),
		EndTime:    TimeFromProto(pb.GetEndTime()),
		AltitudeLo: pb.AltitudeLo,
		AltitudeHi: pb.AltitudeHi,
	}
}
