package store

import (
	"github.com/interuss/dss/pkg/aux_/repos"
	dssstore "github.com/interuss/dss/pkg/store"
)

// aux_.store.Store is a generic means to obtain an aux Repository (repo containing auxiliary
// information not related to standardized services like RID or SCD specifically) to perform
// aux-specific operations on any type of data backing the DSS may ever use.
type Store = dssstore.Store[repos.Repository]
