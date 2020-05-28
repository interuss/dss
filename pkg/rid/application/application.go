package application

import (
	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/rid/repos"
	"go.uber.org/zap"
)

// DefaultClock allows stubbing out the clock for a test clock.
var DefaultClock = clockwork.NewRealClock()

// app contains all of the per-entity Applications.
type app struct {
	// TODO: don't fully embed the repos once we reduce the complexity in the store.
	// Right now it's "coincidence" that the repo has the same signatures as the App interface
	// but we will want to simplify the repos and add the complexity here.
	repos.Transactor
	clock  clockwork.Clock
	logger *zap.Logger
}

type App interface {
	ISAApp
	SubscriptionApp
}

// NewFromTransactor is a convenience function for creating an App
// with the given store.
func NewFromTransactor(transactor repos.Transactor, logger *zap.Logger) App {
	return &app{
		Transactor: transactor,
		clock:      DefaultClock,
		logger:     logger,
	}
}
