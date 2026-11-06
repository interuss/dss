package models

import (
	"strings"
	"testing"
	"time"

	"github.com/interuss/dss/pkg/pb/scdpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

var defaultDate = time.Date(2026, 10, 28, 9, 30, 0, 123456789, time.UTC)

func TestOVNProtoRoundTrip(t *testing.T) {
	uuids := OVN("0b6e0ddb-0b3f-4e1f-8c6e-3b9a3c9f5e2a_019a2b3c-4d5e-7f60-8a1b-2c3d4e5f6a7b")
	digest := NewOVNFromTime(defaultDate, "salt")

	for name, tc := range map[string]struct {
		ovn    OVN
		packed bool
	}{
		"uuids":              {ovn: uuids, packed: true},
		"digest":             {ovn: digest},
		"no OVN phrase":      {ovn: NoOvnPhrase},
		"arbitrary text":     {ovn: OVN("some-client-provided-ovn")},
		"non canonical uuid": {ovn: OVN(strings.ToUpper(string(uuids)))},
	} {
		t.Run(name, func(t *testing.T) {
			pb := OVNToProto(tc.ovn)
			_, isText := pb.GetValue().(*scdpb.OVN_Text)
			assert.Equal(t, !tc.packed, isText)

			buf, err := proto.Marshal(pb)
			require.NoError(t, err)
			if tc.packed {
				assert.Less(t, len(buf), len(tc.ovn))
			}
			var decodedPb scdpb.OVN
			require.NoError(t, proto.Unmarshal(buf, &decodedPb))

			decoded, err := OVNFromProto(&decodedPb)
			require.NoError(t, err)
			assert.Equal(t, tc.ovn, decoded)
		})
	}

	assert.Nil(t, OVNToProto(""))
	empty, err := OVNFromProto(nil)
	require.NoError(t, err)
	assert.Equal(t, OVN(""), empty)
}

func TestOVNsProtoRoundTrip(t *testing.T) {
	ovns := []OVN{NewOVNFromTime(defaultDate, "salt"), "", NoOvnPhrase}

	decoded, err := OVNsFromProto(OVNsToProto(ovns))
	require.NoError(t, err)
	assert.Equal(t, ovns, decoded)

	assert.Nil(t, OVNsToProto(nil))
	decoded, err = OVNsFromProto(nil)
	require.NoError(t, err)
	assert.Nil(t, decoded)
}

func TestOVNProtoErrors(t *testing.T) {
	_, err := OVNFromProto(&scdpb.OVN{Value: &scdpb.OVN_Uuids{Uuids: []byte{1, 2, 3}}})
	assert.Error(t, err)
}
