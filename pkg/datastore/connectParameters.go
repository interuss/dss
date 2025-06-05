package datastore

import (
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

	// ConnectParameters bundles up parameters used for connecting to a CRDB instance.
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
