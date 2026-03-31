package store

import (
	"github.com/interuss/dss/pkg/rid/repos"
	dssstore "github.com/interuss/dss/pkg/store"
)

type Store = dssstore.Store[repos.Repository]
