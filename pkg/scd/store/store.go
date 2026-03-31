package store

import (
	"github.com/interuss/dss/pkg/scd/repos"
	dssstore "github.com/interuss/dss/pkg/store"
)

type Store = dssstore.Store[repos.Repository]
