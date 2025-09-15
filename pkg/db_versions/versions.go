package dbversions

import (
	"os"
	"strconv"
	"strings"

	"github.com/interuss/stacktrace"
)

const (
	Aux = iota
	Rid
	Scd
)

func GetCurrentMajorCRDBSchemaVersion(dbType int) (int64, error) {
	return getCurrentMajorSchemaVersion("crdb/" + strconv.Itoa(dbType) + ".version")
}

func GetCurrentMajorYugabytechemaVersion(dbType int) (int64, error) {
	return getCurrentMajorSchemaVersion("yugabyte/" + strconv.Itoa(dbType) + ".version")
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
