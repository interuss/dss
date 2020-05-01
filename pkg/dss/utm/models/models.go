package models

type (
	// ID models the id of a UTM entity.
	ID string
)

// String returns the string representation of id.
func (id ID) String() string {
	return string(id)
}
