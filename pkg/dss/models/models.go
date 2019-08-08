package models

import (
	"strconv"
	"time"
)

// Convert updatedAt to a string, why not make it smaller
// WARNING: Changing this will cause RMW errors
// 32 is the highest value allowed by strconv
var versionBase = 32

func timestampToVersionString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return strconv.FormatUint(uint64(t.UnixNano()), versionBase)
}
