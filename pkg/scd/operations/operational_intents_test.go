package operations

import (
	"testing"
	"time"

	"github.com/golang/geo/s2"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestOIRPayloadRoundTrip(t *testing.T) {
	start := time.Date(2026, 10, 28, 9, 30, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	lo, hi := float32(10), float32(120)
	params := func() *validOIRParams {
		return &validOIRParams{
			ID:    dssmodels.ID("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a"),
			State: scdmodels.OperationalIntentStateAccepted,
			Volume: &dssmodels.CellsVolume4D{
				Cells:      s2.CellUnion{s2.CellIDFromToken("89c25")},
				StartTime:  &start,
				EndTime:    &end,
				AltitudeLo: &lo,
				AltitudeHi: &hi,
			},
			USSBaseURL: "https://uss.example.com",
			Key:        map[scdmodels.OVN]bool{},
		}
	}

	full := params()
	full.OVN = "previous-ovn"
	full.NewOVN = "0192e4b8-7a00-7000-8000-000000000000_0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a"
	full.State = scdmodels.OperationalIntentStateActivated
	full.SubscriptionID = dssmodels.ID("5d0f0c4e-1f7a-4b0e-9d3c-2a8b6e4f1c7d")
	full.Key = map[scdmodels.OVN]bool{"ovn1": true, "ovn2": true}

	implicit := params()
	implicit.ImplicitSubscription.ID = dssmodels.ID("7a1e2b3c-4d5e-4f60-8a7b-9c0d1e2f3a4b")
	implicit.ImplicitSubscription.BaseURL = "https://sub.example.com"
	implicit.ImplicitSubscription.ForConstraints = true

	for name, request := range map[string]oirPayload{
		"create":          &createOIRPayload{Manager: "uss1", Params: params()},
		"create implicit": &createOIRPayload{Manager: "uss1", Params: implicit},
		"update":          &updateOIRPayload{Manager: "uss1", Params: full},
	} {
		t.Run(name, func(t *testing.T) {
			handler := Registry[request.OperationID()]
			payload, err := handler.EncodePayload(request)
			require.NoError(t, err)
			buf, err := proto.Marshal(payload)
			require.NoError(t, err)
			require.NoError(t, proto.Unmarshal(buf, payload))

			decoded, err := handler.DecodePayload(payload)
			require.NoError(t, err)
			assert.Equal(t, request, decoded)
		})
	}

	_, err := Registry[restapi.CreateOperationalIntentReferenceOperationID].EncodePayload(&createOIRPayload{Manager: "uss1", Params: &validOIRParams{ID: full.ID}})
	require.Error(t, err)
}
