This project is layed out as following:


## Models

Contains the domain objects that are used internally throughout the system, by
application logic and database.

They should contain functions to convert to/from API objects, as well as any
validation logic, and not much else.

## Server

Contains the API for serving the DSS.

Code in the server package should *only*:

1. Convert API objects to domain objects (models)
1. Call any validation functions on the  domain objects (the logic should be present in the models package)
1. Call out to the application layer to perform its task.

## Application

The application layer should be responsible for the bulk of the DSS "logic", and
should call out to the repository layer (currently backed by cockroachdb). Logic 
may include, calling the repo to insert an ISA, then calling the Sub repo to
query for affected Subscriptions.

NOTE: The application layer can be aware of transactions.

## Repository

The repository is simply a set of interfaces that allow us to swap out our 
underlying storage layer.

The repository layer should provide *simple* CRUD style operations. It should
*only* interact with it's given entity. Any cross entity txn's should be done 
via the application layer.

## CockroachDB

The implementation of the Repository layer.


## TODO

There's still a lot to be done here.

1. Add logic to the application layer, and expose transactions in a siimple manner.

Something like adding a `WithTransaction` method to each repo interface, that 
returns a copy of the repo with the cockroach.DB swapped out for a txn. This could
allow:

```
tx := ISARepo.Begin()
ir := ISARepo.WithTransaction(tx)
ir.Search(...)
sr := SubscriptionRepo.WithTransaction(tx)
sr.Insert(...)
tx.Commit()
```

1. Add a new model for each cell table (ISA and Subscriptions). This should have
plumbing for an Application interface, Repo, Store, and everything.

1. Reduce usage of transactions where possible & safe.

1. Add retries! https://github.com/cockroachdb/cockroach-go/tree/master/crdb 

1. The application layers should be responsible for doing the cross queries.

1. Smaller nits
  * Require timestamps in the DB removes nullable time
  * Once GC is implemented we don't need to query for `ends_at > now`, removing
  the need for clockwork in the database layer.
  * AdjusttimeRange can probably be removed.