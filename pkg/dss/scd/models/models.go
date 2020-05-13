package models

import "time"

type (
	// ID models the id of an entity.
	ID string

	// OVN models an opaque version number.
	OVN struct {
		time.Time
	}

	// Version models the version of an entity.
	//
	// Primarily used as a fencing token in data mutations.
	Version int32
)

// String returns the string representation of id.
func (id ID) String() string {
	return string(id)
}

// Empty returns true if ovn indicates an empty opaque version number.
func (ovn OVN) Empty() bool {
	return ovn.IsZero()
}

func (ovn OVN) String() string {
	return ovn.Format(time.RFC3339)
}

// Empty returns true if the value of v indicates an empty version.
func (v Version) Empty() bool {
	return v <= 0
}

// Matches returns true if v matches w.
func (v Version) Matches(w Version) bool {
	return v == w
}
