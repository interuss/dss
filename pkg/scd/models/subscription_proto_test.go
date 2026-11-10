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

func TestSubscriptionProtoRoundTrip(t *testing.T) {
	start := time.Date(2026, 10, 28, 9, 30, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	lo, hi := float32(10), float32(120)

	for name, sub := range map[string]*Subscription{
		"all fields": {
			ID:                          dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a"),
			Manager:                     dssmodels.Manager("uss1"),
			USSBaseURL:                  "https://uss.example.com",
			NotifyForOperationalIntents: true,
			NotifyForConstraints:        true,
			ImplicitSubscription:        true,
			CellsVolume4D: &dssmodels.CellsVolume4D{
				Cells:      s2.CellUnion{s2.CellIDFromToken("89c25"), s2.CellIDFromToken("89c27")},
				StartTime:  &start,
				EndTime:    &end,
				AltitudeLo: &lo,
				AltitudeHi: &hi,
			},
		},
		"empty fields": {
			ID:            dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a"),
			CellsVolume4D: &dssmodels.CellsVolume4D{Cells: s2.CellUnion{}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			pb, err := sub.ToProto()
			require.NoError(t, err)
			buf, err := proto.Marshal(pb)
			require.NoError(t, err)
			var decodedPb scdpb.Subscription
			require.NoError(t, proto.Unmarshal(buf, &decodedPb))

			var decoded Subscription
			require.NoError(t, decoded.FromProto(&decodedPb))
			assert.Equal(t, sub, &decoded)
		})
	}
}

func TestSubscriptionProtoErrors(t *testing.T) {
	var missing *Subscription
	_, err := missing.ToProto()
	require.Error(t, err)

	_, err = (&Subscription{ID: dssmodels.ID("not-a-uuid"), CellsVolume4D: &dssmodels.CellsVolume4D{}}).ToProto()
	require.Error(t, err)

	_, err = (&Subscription{ID: dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a")}).ToProto()
	require.Error(t, err)

	var sub Subscription
	require.Error(t, sub.FromProto(nil))
	require.Error(t, sub.FromProto(&scdpb.Subscription{Id: []byte{1, 2, 3}, Volume: (&dssmodels.CellsVolume4D{}).ToProto()}))
	id, err := dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a").ToProto()
	require.NoError(t, err)
	require.Error(t, sub.FromProto(&scdpb.Subscription{Id: id}))
}
