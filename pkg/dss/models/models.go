package models

import (
	"errors"
	"strconv"
	"time"
)

type (
	ID      string
	Owner   string
	Version struct {
		t time.Time
		s string
	}
	EmptyVersionPolicy int
)

const (
	// Convert updatedAt to a string, why not make it smaller
	// WARNING: Changing this will cause RMW errors
	// 32 is the highest value allowed by strconv
	versionBase                                          = 32
	EmptyVersionPolicyRequireNonEmpty EmptyVersionPolicy = 0
	EmptyVersionPolicyRelaxed         EmptyVersionPolicy = 1
)

func (id ID) String() string {
	return string(id)
}

func (owner Owner) String() string {
	return string(owner)
}

func VersionFromString(s string, evp EmptyVersionPolicy) (*Version, error) {
	v := &Version{s: s}
	if s == "" {
		if evp == EmptyVersionPolicyRequireNonEmpty {
			return nil, errors.New("requires version string")
		}
		return nil, nil
	}
	nanos, err := strconv.ParseUint(string(s), versionBase, 64)
	if err != nil {
		return nil, err
	}
	v.t = time.Unix(0, int64(nanos))
	return v, nil
}

func VersionFromTime(t time.Time) *Version {
	return &Version{
		t: t,
		s: strconv.FormatUint(uint64(t.UnixNano()), versionBase),
	}
}

func (v *Version) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	temp := VersionFromTime(src.(time.Time))
	*v = *temp
	return nil
}

func (v *Version) Empty() bool {
	return v == nil
}

func (v *Version) Matches(v2 *Version) bool {
	if v == nil || v2 == nil {
		return false
	}
	return v.s == v2.s
}

func (v *Version) String() string {
	if v == nil {
		return ""
	}
	return v.s
}

func (v *Version) ToTimestamp() time.Time {
	return v.t
}
