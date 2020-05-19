package cockroach

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildURI(t *testing.T) {
	cases := []struct {
		name   string
		params map[string]string
		want   string
	}{
		{
			name: "valid URI",
			params: map[string]string{
				"host":             "localhost",
				"port":             "26257",
				"user":             "root",
				"ssl_mode":         "enable",
				"ssl_dir":          "/tmp",
				"application_name": "test-app",
			},
			want: "postgresql://root@localhost:26257?application_name=test-app&sslmode=enable&sslrootcert=/tmp/ca.crt&sslcert=/tmp/client.root.crt&sslkey=/tmp/client.root.key",
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
			want: "postgresql://root@localhost:26257?application_name=dss&sslmode=disable",
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
		got, _ := BuildURI(c.params)
		require.Equal(t, c.want, got)
	}
}
