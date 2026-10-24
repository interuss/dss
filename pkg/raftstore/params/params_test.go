package params

import (
	"flag"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeerMap(t *testing.T) {
	tests := []struct {
		name      string
		peers     string
		mutualTLS bool
		want      map[uint64]*url.URL
		wantError bool
	}{
		{
			name:  "valid single peer",
			peers: "1=http://node1:9021",
			want:  map[uint64]*url.URL{1: mustParseURL("http://node1:9021")},
		},
		{
			name:  "valid multiple peers",
			peers: "1=http://node1:9021,2=http://node2:9021,3=http://node3:9021",
			want: map[uint64]*url.URL{
				1: mustParseURL("http://node1:9021"),
				2: mustParseURL("http://node2:9021"),
				3: mustParseURL("http://node3:9021"),
			},
		},
		{
			name:  "valid URL with equals sign in query string",
			peers: "1=http://node1:9021?token=abc123",
			want:  map[uint64]*url.URL{1: mustParseURL("http://node1:9021?token=abc123")},
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
			peers:     "abc=http://node1:9021",
			wantError: true,
		},
		{
			name:      "invalid negative node ID",
			peers:     "-1=http://node1:9021",
			wantError: true,
		},
		{
			name:      "mixed valid and invalid entries",
			peers:     "1=http://node1:9021,badentry",
			wantError: true,
		},
		{
			name:      "invalid zero peer ID",
			peers:     "0=http://node1:9021",
			wantError: true,
		},
		{
			name:      "duplicate peer IDs",
			peers:     "1=http://node1:9021,1=http://node2:9021",
			wantError: true,
		},
		{
			name:      "valid https peers with mutual TLS",
			peers:     "1=https://node1:9021,2=https://node2:9021",
			mutualTLS: true,
			want: map[uint64]*url.URL{
				1: mustParseURL("https://node1:9021"),
				2: mustParseURL("https://node2:9021"),
			},
		},
		{
			name:      "invalid http peer with mutual TLS",
			peers:     "1=https://node1:9021,2=http://node2:9021",
			mutualTLS: true,
			wantError: true,
		},
		{
			name:      "invalid https peer without mutual TLS",
			peers:     "1=http://node1:9021,2=https://node2:9021",
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := ConnectParameters{Peers: tc.peers}
			if tc.mutualTLS {
				c.CertFile, c.KeyFile, c.TrustedCAFile = "node.crt", "node.key", "cas.crt"
			}
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

func TestGetConnectParametersTLSFlags(t *testing.T) {
	setFlag := func(t *testing.T, name string, value string) {
		require.NoError(t, flag.Set(name, value))
		t.Cleanup(func() { require.NoError(t, flag.Set(name, flag.Lookup(name).DefValue)) })
	}

	tests := []struct {
		name          string
		certFile      string
		keyFile       string
		trustedCAFile string
		wantError     bool
	}{
		{name: "none set: no mutual TLS"},
		{name: "all set: mutual TLS", certFile: "node.crt", keyFile: "node.key", trustedCAFile: "cas.crt"},
		{name: "only cert", certFile: "node.crt", wantError: true},
		{name: "only key", keyFile: "node.key", wantError: true},
		{name: "only trusted CAs", trustedCAFile: "cas.crt", wantError: true},
		{name: "cert and key", certFile: "node.crt", keyFile: "node.key", wantError: true},
		{name: "cert and trusted CAs", certFile: "node.crt", trustedCAFile: "cas.crt", wantError: true},
		{name: "key and trusted CAs", keyFile: "node.key", trustedCAFile: "cas.crt", wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setFlag(t, "raft_node_id", "1")
			setFlag(t, "raft_cert_file", tc.certFile)
			setFlag(t, "raft_key_file", tc.keyFile)
			setFlag(t, "raft_trusted_ca_file", tc.trustedCAFile)

			got, err := GetConnectParameters("rid")
			if tc.wantError {
				require.ErrorContains(t, err, "must be set together")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.certFile != "", got.TLSEnabled())
			assert.Equal(t, []string{tc.certFile, tc.keyFile, tc.trustedCAFile}, []string{got.CertFile, got.KeyFile, got.TrustedCAFile})
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
