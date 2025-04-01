package models

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/interuss/stacktrace"
	"github.com/jackc/pgx/v5/pgtype"
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

	// Set a max limit for the SELECT query result
	MaxResultLimit = 10000
)

// PgUUID converts an ID to a pgtype.UUID.
// If the ID this is called on is nil, nil will be returned
func (id *ID) PgUUID() (*pgtype.UUID, error) {
	if id == nil {
		return nil, nil
	}
	pgUUID := pgtype.UUID{}
	err := (&pgUUID).Scan(id.String())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error converting ID to PgUUID format")
	}
	return &pgUUID, nil
}

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

// IDFromString converts a string to an ID, typically provided by an API client.
// The OpenAPI spec for the SCD endpoints defines precisely the regex that UUIDs must conform to:
//
//	^[0-9a-fA-F]{8}\\-[0-9a-fA-F]{4}\\-4[0-9a-fA-F]{3}\\-[8-b][0-9a-fA-F]{3}\\-[0-9a-fA-F]{12}$
//
// This is the strictest possible interpretation for a "V4 UUID RFC 4122 Variant" (See RFC 4122, section-4.4) and means we must refuse
// even UUIDs that have the correct format but are not technically V4 UUID as described RFC 4122 variant (notably the fixed bits)
//
// RFC 4122 is the only RFC that describes a "version 4" UUID: the format requirement using the above regex is
// thus also extended to the RID v1 and v2 endpoints, which explicitly mandate that the UUIDs to be used are the V4 one.
//
// See https://github.com/astm-utm/Protocol/blob/cb7cf962d3a0c01b5ab12502f5f54789624977bf/utm.yaml#L128 for more details
func IDFromString(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ID(""), stacktrace.Propagate(err, "Error parsing ID in UUID V4 format: `%s`", s)
	}

	if id.Variant() != uuid.RFC4122 || id.Version() != 4 {
		return ID(""), stacktrace.NewError("UUID must be V4 as per RFC4122, was: `%v`, id `%s`", id.Variant(), s)
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

	nanos, err := strconv.ParseInt(string(s), versionBase, 64)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing version to integer")
	}
	if nanos < 0 {
		return nil, stacktrace.NewError("Parsed negative value for nanosecond timestamp for version")
	}
	v.t = time.Unix(0, nanos)
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
