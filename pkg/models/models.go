package models

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type (
	// ID represents a UUID string.
	ID string
	// Owner is the owner taken from the oauth token.
	Owner string
	// Version represents a version, which can be supplied as a commit timestamp
	// or a string.
	Version struct {
		t time.Time
		s string
	}
)

const (
	// Convert updatedAt to a string, why not make it smaller
	// WARNING: Changing this will cause RMW errors
	// 32 used to be the highest value allowed by strconv. The new value is 36,
	// although changes to this will result in RMW errors.
	versionBase = 32
)

func (id ID) String() string {
	return string(id)
}

func (owner Owner) String() string {
	return string(owner)
}

// VersionFromString converts a version, typically provided from a user, to
// a Version struct.
func VersionFromString(s string) (*Version, error) {
	if s == "" {
		return nil, errors.New("requires version string")
	}
	v := &Version{s: s}

	nanos, err := strconv.ParseUint(string(s), versionBase, 64)
	if err != nil {
		return nil, err
	}
	v.t = time.Unix(0, int64(nanos))
	return v, nil
}

// VersionFromTime converts a timestamp, typically from the database, to a
// Version struct.
func VersionFromTime(t time.Time) *Version {
	return &Version{
		t: t,
		s: strconv.FormatUint(uint64(t.UnixNano()), versionBase),
	}
}

// Scan implements database/sql's scan interface.
func (v *Version) Scan(src interface{}) error {
	fmt.Println("ERROR SCANNING")
	if src == nil {
		return nil
	}
	fmt.Println(src)
	t, ok := src.(time.Time)
	if !ok {
		panic("not ok@")
	}
	temp := VersionFromTime(t)
	*v = *temp
	fmt.Println("JK no errors scanning")
	return nil
}

// Empty checks if the version is nil.
func (v *Version) Empty() bool {
	return v == nil || v.t.IsZero()
}

// Matches returns true if 2 versions are equal.
func (v *Version) Matches(v2 *Version) bool {
	if v == nil || v2 == nil {
		return false
	}
	return v.s == v2.s
}

// String returns the string representation of a version.
func (v *Version) String() string {
	if v == nil {
		return ""
	}
	return v.s
}

// ToTimestamp converts the version back its commit timestamp.
func (v *Version) ToTimestamp() *time.Time {
	return &v.t
}
