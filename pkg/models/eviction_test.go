package models

import (
	"testing"
	"time"

	"github.com/interuss/dss/pkg/pb/evictionpb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestEvictionParametersFromProto(t *testing.T) {
	testCases := []struct {
		name    string
		pb      *evictionpb.Parameters
		want    EvictionParameters
		wantErr bool
	}{{
		name:    "missing",
		pb:      nil,
		wantErr: true,
	}, {
		name: "present is enabled",
		pb:   &evictionpb.Parameters{Ttl: durationpb.New(time.Minute), Limit: 2},
		want: EvictionParameters{TTL: time.Minute, Limit: 2},
	}, {
		name:    "present without fields is validated",
		pb:      &evictionpb.Parameters{},
		wantErr: true,
	}, {
		name:    "negative TTL",
		pb:      &evictionpb.Parameters{Ttl: durationpb.New(-time.Minute)},
		wantErr: true,
	}, {
		name:    "negative limit",
		pb:      &evictionpb.Parameters{Ttl: durationpb.New(time.Minute), Limit: -1},
		wantErr: true,
	}}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var got EvictionParameters
			err := got.FromProto(testCase.pb)
			if testCase.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, testCase.want, got)
		})
	}
}

func TestEvictionParametersRoundTrip(t *testing.T) {
	p := EvictionParameters{TTL: 30 * time.Minute, Limit: 5}
	var got EvictionParameters
	require.NoError(t, got.FromProto(p.ToProto()))
	require.Equal(t, p, got)
}

func TestEvictionConfigProtoRoundTrip(t *testing.T) {
	for name, c := range map[string]EvictionConfig{
		"scheduled": {Interval: time.Hour, Parameters: EvictionParameters{TTL: time.Minute, Limit: 2}},
	} {
		t.Run(name, func(t *testing.T) {
			var got EvictionConfig
			require.NoError(t, got.FromProto(c.ToProto()))
			require.Equal(t, c, got)
		})
	}

	var got EvictionConfig
	params := (&EvictionParameters{TTL: time.Minute}).ToProto()
	require.Error(t, got.FromProto(&evictionpb.Config{Parameters: params}), "a missing interval is invalid")
	require.Error(t, got.FromProto(&evictionpb.Config{Interval: durationpb.New(0), Parameters: params}), "a zero interval is invalid")
	require.Error(t, got.FromProto(&evictionpb.Config{Interval: durationpb.New(-time.Hour), Parameters: params}), "a negative interval is invalid")
	require.Error(t, got.FromProto(&evictionpb.Config{Interval: durationpb.New(time.Hour)}), "missing parameters are invalid")
}
