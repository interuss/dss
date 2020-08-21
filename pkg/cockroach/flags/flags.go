package flags

import (
	"flag"

	"github.com/interuss/dss/pkg/cockroach"
)

var (
	connectParameters cockroach.ConnectParameters
)

// ConnectParameters returns a ConnectParameters instance that gets populated from well-known CLI flags.
func ConnectParameters() cockroach.ConnectParameters {
	return connectParameters
}

func init() {
	flag.StringVar(&connectParameters.ApplicationName, "cockroach_application_name", "dss", "application name for tagging the connection to cockroach")
	flag.StringVar(&connectParameters.DBName, "cockroach_db_name", "dss", "application name for tagging the connection to cockroach")
	flag.StringVar(&connectParameters.Host, "cockroach_host", "", "cockroach host to connect to")
	flag.IntVar(&connectParameters.Port, "cockroach_port", 26257, "cockroach port to connect to")
	flag.StringVar(&connectParameters.SSL.Mode, "cockroach_ssl_mode", "disable", "cockroach sslmode")
	flag.StringVar(&connectParameters.SSL.Dir, "cockroach_ssl_dir", "", "directory to ssl certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key")
	flag.StringVar(&connectParameters.Credentials.Username, "cockroach_user", "root", "cockroach user to authenticate as")
}
