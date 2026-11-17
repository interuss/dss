package admin

import (
	"time"

	"github.com/interuss/dss/pkg/admin/client"
	"github.com/interuss/stacktrace"
	"github.com/spf13/pflag"
)

var (
	socket  string
	timeout time.Duration
)

// RegisterFlags registers the flags reaching the admin API of the DSS node.
func RegisterFlags(fs *pflag.FlagSet) {
	fs.StringVar(&socket, "admin_socket", "/run/dss/admin.sock", "Path of the Unix socket of the admin API of the DSS node, matching its --admin_socket")
	fs.DurationVar(&timeout, "admin_timeout", time.Minute, "Timeout of each request to the admin API")
}

// NewRaftGroup returns a client of the admin API for the Raft group name.
func NewRaftGroup(name string) (*client.RaftGroup, error) {
	group, err := client.NewRaftGroup(socket, name, timeout)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid --admin_socket")
	}
	return group, nil
}
