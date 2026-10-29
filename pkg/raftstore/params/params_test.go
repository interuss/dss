package params

import (
	"flag"
	"net/url"
	"strconv"
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
			c := ConnectParameters{Peers: tc.peers, Insecure: !tc.mutualTLS}
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

	const (
		required   = "--raft_tls_cert_file, --raft_tls_key_file and --raft_tls_ca_file are required unless --raft_insecure is set"
		notWithTLS = "--raft_insecure cannot be used with --raft_tls_cert_file, --raft_tls_key_file or --raft_tls_ca_file"
	)
	tests := []struct {
		name      string
		certFile  string
		keyFile   string
		caFile    string
		insecure  bool
		wantError string
	}{
		{name: "all TLS flags: mutual TLS", certFile: "node.crt", keyFile: "node.key", caFile: "cas.crt"},
		{name: "insecure only: plain HTTP", insecure: true},
		{name: "nothing set", wantError: required},
		{name: "all TLS flags and insecure", certFile: "node.crt", keyFile: "node.key", caFile: "cas.crt", insecure: true, wantError: notWithTLS},
		{name: "only cert", certFile: "node.crt", wantError: required},
		{name: "only key", keyFile: "node.key", wantError: required},
		{name: "only CA file", caFile: "cas.crt", wantError: required},
		{name: "cert and key", certFile: "node.crt", keyFile: "node.key", wantError: required},
		{name: "cert and CA file", certFile: "node.crt", caFile: "cas.crt", wantError: required},
		{name: "key and CA file", keyFile: "node.key", caFile: "cas.crt", wantError: required},
		{name: "only cert and insecure", certFile: "node.crt", insecure: true, wantError: notWithTLS},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setFlag(t, "raft_node_id", "1")
			setFlag(t, "raft_peers", "1=http://127.0.0.1:9011")
			setFlag(t, "raft_tls_cert_file", tc.certFile)
			setFlag(t, "raft_tls_key_file", tc.keyFile)
			setFlag(t, "raft_tls_ca_file", tc.caFile)
			setFlag(t, "raft_insecure", strconv.FormatBool(tc.insecure))

			got, err := GetConnectParameters("rid")
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.certFile != "", got.TLSEnabled())
			assert.Equal(t, []string{tc.certFile, tc.keyFile, tc.caFile}, []string{got.TLSCertFile, got.TLSKeyFile, got.TLSCAFile})
		})
	}
}

func TestGroupClusterID(t *testing.T) {
	ids := map[uint64]string{}
	for _, clusterID := range []uint64{1, 2} {
		for _, group := range []string{"aux", "rid", "scd"} {
			id, err := groupClusterID(clusterID, group)
			require.NoError(t, err)
			again, err := groupClusterID(clusterID, group)
			require.NoError(t, err)
			assert.Equal(t, id, again)
			assert.NotContains(t, ids, id)
			ids[id] = group
		}
	}
}

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return u
}
