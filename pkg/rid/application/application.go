package application

// App contains all of the per-entity Applications.
type App struct {
	ISA          ISAAppInterface
	Subscription SubscriptionAppInterface
}
