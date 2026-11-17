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
)

// RegisterFlags registers the flags reaching the admin API of the DSS node.
func RegisterFlags(fs *pflag.FlagSet) {
	fs.StringVar(&Socket, "admin_socket", "/run/dss/admin.sock", "Path of the Unix socket of the admin API of the DSS node, matching its --admin_socket")
	fs.DurationVar(&Timeout, "admin_timeout", time.Minute, "Timeout of each request to the admin API")
}
