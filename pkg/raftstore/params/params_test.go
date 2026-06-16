package params

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeerMap(t *testing.T) {
	tests := []struct {
		name      string
		peers     string
		want      map[uint64]*url.URL
		wantError bool
	}{
		{
			name:  "valid single peer",
			peers: "1=https://node1:9021",
			want:  map[uint64]*url.URL{1: mustParseURL("https://node1:9021")},
		},
		{
			name:  "valid multiple peers",
			peers: "1=https://node1:9021,2=https://node2:9021,3=https://node3:9021",
			want: map[uint64]*url.URL{
				1: mustParseURL("https://node1:9021"),
				2: mustParseURL("https://node2:9021"),
				3: mustParseURL("https://node3:9021"),
			},
		},
		{
			name:  "valid URL with equals sign in query string",
			peers: "1=https://node1:9021?token=abc123",
			want:  map[uint64]*url.URL{1: mustParseURL("https://node1:9021?token=abc123")},
		},
		{
			name:      "invalid empty peers string",
			peers:     "",
			wantError: true,
		},
		{
			name:      "invalid entry format",
			peers:     "invalidentry",
			wantError: true,
		},
		{
			name:      "invalid non-numeric node ID",
			peers:     "abc=https://node1:9021",
			wantError: true,
		},
		{
			name:      "invalid negative node ID",
			peers:     "-1=https://node1:9021",
			wantError: true,
		},
		{
			name:      "mixed valid and invalid entries",
			peers:     "1=https://node1:9021,badentry",
			wantError: true,
		},
		{
			name:      "invalid zero peer ID",
			peers:     "0=https://node1:9021",
			wantError: true,
		},
		{
			name:      "duplicate peer IDs",
			peers:     "1=https://node1:9021,1=https://node2:9021",
			wantError: true,
		},
		{
			name:      "invalid URL scheme",
			peers:     "1=http://node1:9021",
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ConnectParameters{Peers: tc.peers}
			got, err := c.PeerMap()
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return u
}
