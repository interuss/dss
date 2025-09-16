package dbversions

import (
	_ "embed"
	"strconv"
	"strings"

	"github.com/interuss/stacktrace"
)

// Embedded version files
//
//go:embed crdb/aux_.version
var crdbAuxVersion string

//go:embed crdb/rid.version
var crdbRidVersion string

//go:embed crdb/scd.version
var crdbScdVersion string

//go:embed yugabyte/aux_.version
var yugabyteAuxVersion string

//go:embed yugabyte/rid.version
var yugabyteRidVersion string

//go:embed yugabyte/scd.version
var yugabyteScdVersion string

const (
	Aux = "aux"
	Rid = "rid"
	Scd = "scd"
)

var crdbVersions = map[string]string{
	Aux: crdbAuxVersion,
	Rid: crdbRidVersion,
	Scd: crdbScdVersion,
}

var yugabyteVersions = map[string]string{
	Aux: yugabyteAuxVersion,
	Rid: yugabyteRidVersion,
	Scd: yugabyteScdVersion,
}

// getMajorVersion parses and returns the major version from a version string.
func getMajorVersion(version string) (int64, error) {
	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) == 0 {
		return 0, stacktrace.NewError("Invalid version string: %s", version)
	}

	v, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to convert major version from %s", version)
	}
	return int64(v), nil
}

// GetCurrentMajorCRDBSchemaVersion returns the major schema version for a CRDB type.
func GetCurrentMajorCRDBSchemaVersion(dbType string) (int64, error) {
	version, ok := crdbVersions[dbType]
	if !ok {
		return 0, stacktrace.NewError("db type %s does not exist", dbType)
	}
	return getMajorVersion(version)
}

// GetCurrentMajorYugabyteSchemaVersion returns the major schema version for a Yugabyte type.
func GetCurrentMajorYugabyteSchemaVersion(dbType string) (int64, error) {
	version, ok := yugabyteVersions[dbType]
	if !ok {
		return 0, stacktrace.NewError("db type %s does not exist", dbType)
	}
	return getMajorVersion(version)
}
