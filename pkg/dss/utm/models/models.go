package models

type (
	// ID models the id of a UTM entity.
	ID string
	// Owner models the owner of a UTM entity.
	Owner string
	// Subscription models a UTM subscription.
	Subscription struct{}
)

// String returns the string representation of id.
func (id ID) String() string {
	return string(id)
}

// String returns the string representation of owner.
func (owner Owner) String() string {
	return string(owner)
}
