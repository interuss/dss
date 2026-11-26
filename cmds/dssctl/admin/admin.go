package admin

import (
	"time"

	"github.com/spf13/pflag"
)

var (
	// Socket is the path of the Unix socket of the admin API of the DSS node.
	Socket string
	// Timeout is the timeout of each request to the admin API.
	Timeout time.Duration
	// DumpRequests prints the full requests to and responses from the admin API.
	DumpRequests bool
)

// RegisterFlags registers the flags reaching the admin API of the DSS node.
func RegisterFlags(fs *pflag.FlagSet) {
	fs.StringVar(&Socket, "admin_socket", "/run/dss/admin.sock", "Path of the Unix socket of the admin API of the DSS node, matching its --admin_socket")
	fs.DurationVar(&Timeout, "admin_timeout", time.Minute, "Timeout of each request to the admin API")
	fs.BoolVar(&DumpRequests, "admin_dump_requests", false, "Print the full admin requests and responses to stderr")
}
