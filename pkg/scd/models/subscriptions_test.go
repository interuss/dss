package models

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestValidateDependentOpAltitudes(t *testing.T) {
	cells := s2.CellUnion{s2.CellID(8768904281496485888)}
	start := time.Date(2026, time.November, 6, 16, 26, 30, 0, time.UTC)
	end := start.Add(time.Hour)

	op := &OperationalIntent{
		ID: "00000185-e36d-40be-8d38-beca6ca31002",
		CellsVolume4D: &dssmodels.CellsVolume4D{
			Cells:      cells,
			StartTime:  &start,
			EndTime:    &end,
			AltitudeLo: new(float32(10)),
			AltitudeHi: new(float32(50)),
		},
	}

	testCases := []struct {
		name    string
		lo, hi  *float32
		wantErr bool
	}{
		{name: "no bounds"},
		{name: "covering bounds", lo: new(float32(10)), hi: new(float32(50))},
		{name: "only lower bound, covering", lo: new(float32(10))},
		{name: "only upper bound, covering", hi: new(float32(50))},
		{name: "lower bound above intent", lo: new(float32(11)), hi: new(float32(50)), wantErr: true},
		{name: "upper bound below intent", lo: new(float32(10)), hi: new(float32(49)), wantErr: true},
		{name: "only lower bound, above intent", lo: new(float32(11)), wantErr: true},
		{name: "only upper bound, below intent", hi: new(float32(49)), wantErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sub := &Subscription{
				CellsVolume4D: &dssmodels.CellsVolume4D{
					Cells:      cells,
					StartTime:  &start,
					EndTime:    &end,
					AltitudeLo: testCase.lo,
					AltitudeHi: testCase.hi,
				},
			}

			err := sub.ValidateDependentOp(op)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
