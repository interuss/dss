package dbversions

import (
	"os"
	"strconv"
	"strings"

	"github.com/interuss/stacktrace"
)

const (
	Aux = "aux"
	Rid = "rid"
	Scd = "scd"
)

func GetCurrentMajorCRDBSchemaVersion(dbType string) (int64, error) {
	return getCurrentMajorSchemaVersion("crdb/" + dbType + ".version")
}

func GetCurrentMajorYugabytechemaVersion(dbType string) (int64, error) {
	return getCurrentMajorSchemaVersion("yugabyte/" + dbType + ".version")
}

func getCurrentMajorSchemaVersion(file string) (int64, error) {
	buf, err := os.ReadFile(file)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to read schema version file '%s'", file)
	}

	v, err := strconv.Atoi(strings.Split(string(buf), "")[0])
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to convert schema version '%s' to int", string(buf))
	}

	return int64(v), nil
}
