package actions

import (
	"github.com/interuss/dss/pkg/aux_/repos"
	dssstore "github.com/interuss/dss/pkg/store"
)

// Registry maps operation IDs to their handlers.
// TODO: implement
var Registry = map[string]dssstore.OperationHandler[repos.Repository]{}
