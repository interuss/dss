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
	flag.IntVar(&connectParameters.MaxOpenConns, "max_open_conns", 4, "maximum number of open connections to the database, default is 4")
	flag.IntVar(&connectParameters.MaxConnIdleSeconds, "max_conn_idle_secs", 30, "maximum amount of time in seconds a connection may be idle, default is 30 seconds")
	flag.IntVar(&connectParameters.MaxRetries, "cockroach_max_retries", 100, "maximum number of attempts to retry a query in case of contention, default is 100")
}
