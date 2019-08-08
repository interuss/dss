package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

// Convert updatedAt to a string, why not make it smaller
// WARNING: Changing this will cause RMW errors
// 32 is the highest value allowed by strconv
var versionBase = 32

func versionStringToTimestamp(s string) (time.Time, error) {
	var t time.Time
	nanos, err := strconv.ParseUint(s, versionBase, 64)
	if err != nil {
		return t, err
	}
	return time.Unix(0, int64(nanos)), nil
}

func timestampToVersionString(t time.Time) string {
	return strconv.FormatUint(uint64(t.UnixNano()), versionBase)
}

type NullTime struct {
	Time  time.Time
	Valid bool // Valid indicates whether Time carries a non-NULL value.
}

func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time = time.Time{}
		nt.Valid = false
		return nil
	}

	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("failed to cast database value, expected time.Time, got %T", value)
	}
	nt.Time, nt.Valid = t, ok

	return nil
}

func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}
