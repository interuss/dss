package cockroach

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBuildDSN(t *testing.T) {
	cases := []struct {
		name   string
		params map[string]string
		want   string
	}{
		{
			name: "valid URI",
			params: map[string]string{
				"host":     "localhost",
				"port":     "26257",
				"user":     "root",
				"ssl_mode": "enable",
				"ssl_dir":  "/tmp",
			},
			want: "application_name=dss host=localhost pool_max_conns=4 port=26257 sslcert=/tmp/client.root.crt sslkey=/tmp/client.root.key sslmode=enable sslrootcert=/tmp/ca.crt user=root",
		},
		{
			name: "missing host",
			params: map[string]string{
				"port":     "26257",
				"user":     "root",
				"ssl_mode": "enable",
				"ssl_dir":  "/tmp",
			},
			want: "",
		},
		{
			name: "missing port",
			params: map[string]string{
				"host":     "localhost",
				"user":     "root",
				"ssl_mode": "enable",
				"ssl_dir":  "/tmp",
			},
			want: "",
		},
		{
			name: "missing user",
			params: map[string]string{
				"host":     "localhost",
				"port":     "26257",
				"ssl_mode": "enable",
				"ssl_dir":  "/tmp",
			},
			want: "",
		},
		{
			name: "missing ssl_mode",
			params: map[string]string{
				"host":    "localhost",
				"port":    "26257",
				"user":    "root",
				"ssl_dir": "/tmp",
			},
			want: "",
		},
		{
			name: "ssl_disabled",
			params: map[string]string{
				"host":     "localhost",
				"port":     "26257",
				"user":     "root",
				"ssl_mode": "disable",
			},
			want: "application_name=dss host=localhost pool_max_conns=4 port=26257 sslmode=disable user=root",
		},
		{
			name: "missing ssl_dir",
			params: map[string]string{
				"host":     "localhost",
				"port":     "26257",
				"user":     "root",
				"ssl_mode": "enable",
			},
			want: "",
		},
	}
	for _, c := range cases {
		got, _ := connectParametersFromMap(c.params).BuildDSN()
		require.Equal(t, c.want, got)
	}
}

func TestFormatDSN(t *testing.T) {
	params := map[string]string{
		"keyA": "valueA",
		"keyB": "valueB",
	}
	require.Equal(t, "keyA=valueA keyB=valueB", formatDSN(params))
}
