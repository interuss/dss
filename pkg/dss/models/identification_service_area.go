package models

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/steeling/InterUSS-Platform/pkg/dssv1"
)

type IdentificationServiceArea struct {
	S2Model
}

func (i *IdentificationServiceArea) ToProto() (*dssv1.IdentificationServiceArea, error) {
	result := &dssv1.IdentificationServiceArea{
		Id:         i.ID.String(),
		Owner:      i.Owner.String(),
		FlightsUrl: i.Url,
		Version:    i.Version.String(),
	}

	if i.StartTime != nil {
		ts, err := ptypes.TimestampProto(*i.StartTime)
		if err != nil {
			return nil, err
		}
		result.TimeStart = ts
	}

	if i.EndTime != nil {
		ts, err := ptypes.TimestampProto(*i.EndTime)
		if err != nil {
			return nil, err
		}
		result.TimeEnd = ts
	}
	return result, nil
}
