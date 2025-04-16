package flags

import (
	"flag"

	"github.com/interuss/dss/pkg/datastore"
)

var (
	connectParameters datastore.ConnectParameters
)

// ConnectParameters returns a ConnectParameters instance that gets populated from well-known CLI flags.
func ConnectParameters() datastore.ConnectParameters {
	return connectParameters
}

func init() {
	flag.StringVar(&connectParameters.ApplicationName, "database_application_name", "dss", "application name for tagging the connection to the datastore")
	flag.StringVar(&connectParameters.DBName, "datastore_db_name", "dss", "database name within the datastore")
	flag.StringVar(&connectParameters.Host, "datastore_host", "", "datastore host to connect to")
	flag.IntVar(&connectParameters.Port, "datastore_port", 26257, "datastore port to connect to")
	flag.StringVar(&connectParameters.SSL.Mode, "datastore_ssl_mode", "disable", "datastore SSL mode ('enable' or 'disable')")
	flag.StringVar(&connectParameters.SSL.Dir, "datastore_ssl_dir", "", "directory to SSL certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key")
	flag.StringVar(&connectParameters.Credentials.Username, "datastore_user", "root", "datastore user to authenticate as")
	flag.IntVar(&connectParameters.MaxOpenConns, "max_open_conns", 4, "maximum number of open connections to the database, default is 4")
	flag.IntVar(&connectParameters.MaxConnIdleSeconds, "max_conn_idle_secs", 30, "maximum amount of time in seconds a connection may be idle, default is 30 seconds")
	flag.IntVar(&connectParameters.MaxRetries, "datastore_max_retries", 100, "maximum number of attempts to retry a query in case of contention, default is 100")
}
