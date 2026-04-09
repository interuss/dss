package params

import (
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/interuss/stacktrace"
)

type (
	// Credentials models connect credentials.
	Credentials struct {
		Username string
		Password string
	}

	// SSL models SSL configuration parameters.
	SSL struct {
		Mode string
		Dir  string
	}

	// ConnectParameters bundles up parameters used for connecting to a datastore instance.
	ConnectParameters struct {
		ApplicationName    string
		Host               string
		Port               int
		DBName             string
		Credentials        Credentials
		SSL                SSL
		MaxOpenConns       int
		MaxConnIdleSeconds int
		MaxRetries         int
	}
)

var (
	connectParameters ConnectParameters
)

func init() {
	flag.StringVar(&connectParameters.ApplicationName, "datastore_application_name", "dss", "application name for tagging the connection to the database")
	flag.StringVar(&connectParameters.Host, "datastore_host", "", "database host to connect to")
	flag.IntVar(&connectParameters.Port, "datastore_port", 26257, "database port to connect to")
	flag.StringVar(&connectParameters.SSL.Mode, "datastore_ssl_mode", "disable", "database sslmode")
	flag.StringVar(&connectParameters.SSL.Dir, "datastore_ssl_dir", "", "directory to ssl certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key")
	flag.StringVar(&connectParameters.Credentials.Username, "datastore_user", "root", "database user to authenticate as")
	flag.IntVar(&connectParameters.MaxOpenConns, "datastore_max_open_conns", 4, "maximum number of open connections to the database, default is 4")
	flag.IntVar(&connectParameters.MaxConnIdleSeconds, "datastore_max_conn_idle_secs", 30, "maximum amount of time in seconds a connection may be idle, default is 30 seconds")
	flag.IntVar(&connectParameters.MaxRetries, "datastore_max_retries", 100, "maximum number of attempts to retry a query in case of contention, default is 100")

	flag.StringVar(&connectParameters.ApplicationName, "cockroach_application_name", connectParameters.ApplicationName, "DEPRECATED: use 'datastore_application_name' instead")
	flag.StringVar(&connectParameters.Host, "cockroach_host", connectParameters.Host, "DEPRECATED: use 'datastore_host' instead")
	flag.IntVar(&connectParameters.Port, "cockroach_port", connectParameters.Port, "DEPRECATED: use 'datastore_port' instead")
	flag.StringVar(&connectParameters.SSL.Mode, "cockroach_ssl_mode", connectParameters.SSL.Mode, "DEPRECATED: use 'datastore_ssl_mode' instead")
	flag.StringVar(&connectParameters.SSL.Dir, "cockroach_ssl_dir", connectParameters.SSL.Dir, "DEPRECATED: use 'datastore_ssl_dir' instead")
	flag.StringVar(&connectParameters.Credentials.Username, "cockroach_user", connectParameters.Credentials.Username, "DEPRECATED: use 'datastore_user' instead")
	flag.IntVar(&connectParameters.MaxOpenConns, "max_open_conns", connectParameters.MaxOpenConns, "DEPRECATED: use 'datastore_max_open_conns' instead")
	flag.IntVar(&connectParameters.MaxConnIdleSeconds, "max_conn_idle_secs", connectParameters.MaxConnIdleSeconds, "DEPRECATED: use 'datastore_max_conn_idle_secs' instead")
	flag.IntVar(&connectParameters.MaxRetries, "cockroach_max_retries", connectParameters.MaxRetries, "DEPRECATED: use 'datastore_max_retries' instead")
}

// ConnectParameters returns a ConnectParameters instance that gets populated from well-known CLI flags.
func GetConnectParameters() ConnectParameters {
	return connectParameters
}

func parseIntOrDefault(port string, defaultPort int64) int64 {
	p, err := strconv.ParseInt(port, 10, 16)
	if err != nil {
		p = defaultPort
	}
	return p
}

// formatDSN constructs a DSN string from a key value map.
func formatDSN(dsnMap map[string]string) string {
	d := make([]string, 0)
	for key, value := range dsnMap {
		if value != "" {
			d = append(d, fmt.Sprintf("%s=%s", key, value))
		}
	}
	sort.Strings(d)
	return strings.Join(d, " ")
}

// BuildURI returns a URI built from p.
func (cp ConnectParameters) BuildDSN() (string, error) {
	dsnMap := make(map[string]string)

	u := cp.Credentials.Username
	if u == "" {
		return "", stacktrace.NewError("Missing datastore username")
	}
	dsnMap["user"] = u

	h := cp.Host
	if h == "" {
		return "", stacktrace.NewError("Missing datastore hostname")
	}
	dsnMap["host"] = h

	port := cp.Port
	if port == 0 {
		return "", stacktrace.NewError("Missing datastore port")
	}
	dsnMap["port"] = fmt.Sprintf("%d", port)

	an := cp.ApplicationName
	if an == "" {
		an = "dss"
	}
	dsnMap["application_name"] = an

	dbn := cp.DBName
	if dbn != "" {
		dsnMap["dbname"] = dbn
	}

	sslMode := cp.SSL.Mode
	if sslMode == "" {
		return "", stacktrace.NewError("Missing datastore ssl_mode")
	}
	dsnMap["sslmode"] = sslMode

	dsnMap["pool_max_conns"] = fmt.Sprintf("%d", cp.MaxOpenConns)

	if sslMode == "disable" {
		return formatDSN(dsnMap), nil
	}

	dir := cp.SSL.Dir
	if dir == "" {
		return "", stacktrace.NewError("Missing datastore ssl_dir")
	}
	dsnMap["sslrootcert"] = cp.GetCAFile()
	dsnMap["sslcert"] = fmt.Sprintf("%s/client.%s.crt", dir, u)
	dsnMap["sslkey"] = fmt.Sprintf("%s/client.%s.key", dir, u)

	return formatDSN(dsnMap), nil
}

// Return the CA file to use
func (cp ConnectParameters) GetCAFile() string {

	if cp.SSL.Mode == "disable" || cp.SSL.Dir == "" {
		return ""
	}

	return fmt.Sprintf("%s/ca.crt", cp.SSL.Dir)
}

// Return the instance CA file to use
func (cp ConnectParameters) GetInstanceCAFile() string {

	if cp.SSL.Mode == "disable" || cp.SSL.Dir == "" {
		return ""
	}

	return fmt.Sprintf("%s/ca-instance.crt", cp.SSL.Dir)
}
