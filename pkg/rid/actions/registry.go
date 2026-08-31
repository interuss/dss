package actions

import (
	"github.com/interuss/dss/pkg/rid/repos"
	dssstore "github.com/interuss/dss/pkg/store"
)

// Registry maps operation IDs to their handlers.
var Registry = map[string]dssstore.OperationHandler[repos.Repository]{}
