package models

import (
	"strconv"
	"time"
)

const (
	VersionEmpty Version = ""
	// Convert updatedAt to a string, why not make it smaller
	// WARNING: Changing this will cause RMW errors
	// 32 is the highest value allowed by strconv
	versionBase = 32
)

type (
	ID      string
	Owner   string
	Version string
)

func (id ID) String() string {
	return string(id)
}

func (owner Owner) String() string {
	return string(owner)
}

func VersionFromTimestamp(t *time.Time) Version {
	if t == nil {
		return VersionEmpty
	}

	return Version(strconv.FormatUint(uint64(t.UnixNano()), versionBase))
}

func (v Version) String() string {
	return string(v)
}

func (v Version) ToTimestamp() (time.Time, error) {
	var t time.Time
	if v == "" {
		return t, nil
	}
	nanos, err := strconv.ParseUint(string(v), versionBase, 64)
	if err != nil {
		return t, err
	}
	return time.Unix(0, int64(nanos)), nil
}

func ptrToFloat32(f float32) *float32 {
	return &f
}
