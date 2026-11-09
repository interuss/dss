package params

import (
	"flag"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConnectParametersTLSFlags(t *testing.T) {
	setFlag := func(t *testing.T, name string, value string) {
		require.NoError(t, flag.Set(name, value))
		t.Cleanup(func() { require.NoError(t, flag.Set(name, flag.Lookup(name).DefValue)) })
	}

	const (
		required   = "--raft_tls_cert_file, --raft_tls_key_file and --raft_tls_ca_file are required unless --raft_insecure is set"
		notWithTLS = "--raft_insecure cannot be used with --raft_tls_cert_file, --raft_tls_key_file or --raft_tls_ca_file"
		badScheme  = "invalid --raft_peer_url"
	)
	tests := []struct {
		name      string
		url       string
		certFile  string
		keyFile   string
		caFile    string
		insecure  bool
		wantError string
	}{
		{name: "http URL with mutual TLS", url: "http://127.0.0.1:9011", certFile: "node.crt", keyFile: "node.key", caFile: "cas.crt", wantError: badScheme},
		{name: "https URL without mutual TLS", url: "https://127.0.0.1:9011", insecure: true, wantError: badScheme},
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
			nodeURL := tc.url
			if nodeURL == "" {
				nodeURL = "https://127.0.0.1:9011"
				if tc.insecure {
					nodeURL = "http://127.0.0.1:9011"
				}
			}
			setFlag(t, "raft_node_id", "1")
			setFlag(t, "raft_peer_url", nodeURL)
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
