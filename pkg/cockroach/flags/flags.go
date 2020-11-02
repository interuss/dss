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
	flag.IntVar(&connectParameters.MaxOpenConns, "max_open_conns", 0, "maximum number of open connections to the database,  default is 0 (unlimited)")
	flag.IntVar(&connectParameters.MaxIdleConns, "max_idle_conns", 1, "maximum number of connections in the idle connection pool")
	flag.IntVar(&connectParameters.MaxConnLifeSeconds, "max_conn_life_secs", 15, "maximum amount of time in seconds a connection may be reused, default is 15 seconds")
}
