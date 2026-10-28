package models

import (
	"testing"

	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUssAvailabilityStateProtoRoundTrip(t *testing.T) {
	for _, state := range []UssAvailabilityState{UssAvailabilityStateUnknown, UssAvailabilityStateNormal, UssAvailabilityStateDown} {
		t.Run(string(state), func(t *testing.T) {
			pb, err := state.ToProto()
			require.NoError(t, err)
			var decoded UssAvailabilityState
			require.NoError(t, decoded.FromProto(pb))
			assert.Equal(t, state, decoded)
		})
	}
}

func TestUssAvailabilityStateProtoErrors(t *testing.T) {
	_, err := UssAvailabilityState("").ToProto()
	require.Error(t, err)
	_, err = UssAvailabilityState("Busy").ToProto()
	require.Error(t, err)

	var state UssAvailabilityState
	require.Error(t, state.FromProto(scdpb.UssAvailabilityState(42)))
}
