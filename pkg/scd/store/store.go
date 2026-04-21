package store

import (
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
)

// scd.store.Store is a generic means to obtain an SCD scd.repos.Repository to perform SCD-specific
// operations on any type of data backing the DSS may ever use.
type Store = dssstore.Store[repos.Repository]
