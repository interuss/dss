package models

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/interuss/stacktrace"
)

type (
	// ID represents a UUID-formatted string.
	ID string

	// Owner is the owner taken from the oauth token.
	Owner string

	// Semantically similar to Owner
	Manager string

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

func (id ID) Empty() bool {
	return string(id) == ""
}

func (owner Owner) String() string {
	return string(owner)
}

func (manager Manager) String() string {
	return string(manager)
}

func IDFromString(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ID(""), stacktrace.Propagate(err, "Error parsing ID in UUID format")
	}
	return ID(id.String()), nil
}

func ManagerFromString(s string) Manager {
	return Manager(s)
}

func IDFromOptionalString(s string) (ID, error) {
	if s == "" {
		return ID(""), nil
	}
	return IDFromString(s)
}

// VersionFromString converts a version, typically provided from a user, to
// a Version struct.
func VersionFromString(s string) (*Version, error) {
	if s == "" {
		return nil, stacktrace.NewError("Missing version string")
	}
	v := &Version{s: s}

	nanos, err := strconv.ParseUint(string(s), versionBase, 64)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing version to integer")
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
	if src == nil {
		return nil
	}
	t, ok := src.(time.Time)
	if !ok {
		return stacktrace.NewError("Error scanning version")
	}
	temp := VersionFromTime(t)
	*v = *temp
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
	if v == nil {
		t := time.Time{}
		return &t
	}
	return &v.t
}
