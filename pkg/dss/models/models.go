package models

import (
	"strconv"
	"time"
)

// Convert updatedAt to a string, why not make it smaller
// WARNING: Changing this will cause RMW errors
// 32 is the highest value allowed by strconv
var versionBase = 32

func VersionStringToTimestamp(s string) (time.Time, error) {
	var t time.Time
	nanos, err := strconv.ParseUint(s, versionBase, 64)
	if err != nil {
		return t, err
	}
	return time.Unix(0, int64(nanos)), nil
}

func TimestampToVersionString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return strconv.FormatUint(uint64(t.UnixNano()), versionBase)
}
