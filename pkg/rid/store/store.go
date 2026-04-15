package store

import (
	"github.com/interuss/dss/pkg/rid/repos"
	dssstore "github.com/interuss/dss/pkg/store"
)

// rid.store.Store is a generic means to obtain an RID rid.repos.Repository to perform RID-specific
// operations on any type of data backing the DSS may ever use.
type Store = dssstore.Store[repos.Repository]
