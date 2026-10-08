package consensus

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func testProposal() Proposal {
	return Proposal{
		Header: Header{
			ID:          uuid.New(),
			NodeID:      42,
			Timestamp:   time.Date(2026, time.October, 8, 12, 30, 15, 123456789, time.UTC),
			RequestType: "scdv1.CreateOperationalIntentReference",
		},
		Value: []byte(`{"some":"payload"}`),
	}
}

func TestProposalEncodeDecodeRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name   string
		modify func(*Proposal)
	}{
		{name: "write", modify: func(*Proposal) {}},
		{name: "read only", modify: func(p *Proposal) { p.ReadOnly = true }},
		{name: "empty value", modify: func(p *Proposal) { p.Value = []byte{} }},
		{name: "max request type length", modify: func(p *Proposal) { p.RequestType = strings.Repeat("a", 255) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := testProposal()
			tc.modify(&p)

			buf, err := p.encode()
			require.NoError(t, err)
			require.Len(t, buf, 1+headerFixedSize+len(p.RequestType)+len(p.Value))

			decoded, err := decodeProposal(buf)
			require.NoError(t, err)
			require.Equal(t, p, decoded)
		})
	}
}

func TestProposalEncodeErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		modify func(*Proposal)
	}{
		{name: "missing request type", modify: func(p *Proposal) { p.RequestType = "" }},
		{name: "request type too long", modify: func(p *Proposal) { p.RequestType = strings.Repeat("a", 256) }},
		{name: "zero timestamp", modify: func(p *Proposal) { p.Timestamp = time.Time{} }},
		{name: "timestamp after 2262", modify: func(p *Proposal) { p.Timestamp = time.Date(2300, time.January, 1, 0, 0, 0, 0, time.UTC) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := testProposal()
			tc.modify(&p)

			_, err := p.encode()
			require.Error(t, err)
		})
	}
}

func TestProposalDecodeErrors(t *testing.T) {
	valid, err := testProposal().encode()
	require.NoError(t, err)

	withByte := func(i int, b byte) []byte {
		buf := append([]byte{}, valid...)
		buf[i] = b
		return buf
	}

	for _, tc := range []struct {
		name string
		data []byte
	}{
		{name: "empty", data: nil},
		{name: "unsupported version", data: withByte(0, proposalEncodingVersion+1)},
		{name: "truncated fixed header", data: valid[:headerFixedSize]},
		{name: "truncated request type", data: valid[:1+headerFixedSize+1]},
		{name: "unknown flag", data: withByte(1+32, 1<<7)},
		{name: "missing request type", data: withByte(1+33, 0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeProposal(tc.data)
			require.Error(t, err)
		})
	}
}
