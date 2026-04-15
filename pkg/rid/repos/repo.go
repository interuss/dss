package repos

// rid.repos.Repository contains all of the repo interfaces to perform RID operations on any data
// backing.  This is a repository type, generally intended to be obtained/used via a
// store.Store[Repository] interface.
type Repository interface {
	ISA
	Subscription
}
