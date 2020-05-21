package application

import (
	"github.com/dpjacques/clockwork"
	"github.com/interuss/dss/pkg/rid/cockroach"
)

// DefaultClock allows stubbing out the clock for a test clock.
var DefaultClock = clockwork.NewRealClock()

// App contains all of the per-entity Applications.
type App struct {
	ISA          ISAAppInterface
	Subscription SubscriptionAppInterface
}

// NewFromRepo is a convenience function for creating an App
// with the given store.
func NewFromRepo(repo *cockroach.Store) *App {
	return &App{
		ISA:          &ISAApp{repo.ISA, DefaultClock},
		Subscription: &SubscriptionApp{repo.Subscription, DefaultClock},
	}
}
