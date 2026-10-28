# Database cleanup

Data will accumulate over time in the database if client USSs do not remove their expired entities, and this can lead to lower performance due to quantity of stale entities.  For this reason, InterUSS recommends periodically cleaning up no-longer-relevant entities if USS clients do not always clean up after themselves.

## Why and when to run the cleanup

Expired entities (operational intents past their end time, stale subscriptions, ISAs whose lifetime has elapsed) can accumulate over time in the datastore if not cleaned up by clients. While the DSS keeps functioning correctly with some stale rows present, excessive accumulation can impact:

- Storage growth: unbounded storage usage in data stores (e.g., CockroachDB / YugabyteDB).
- Query performance: indexes get larger, range scans on `operational_intents` / `subscriptions` / `identification_service_areas` degrade.

There is no single correct interval or TTL. Reasonable values depend on context and must be defined per DSS pool, taking into account: traffic volume and entity churn, regulatory or contractual data-retention requirements applicable to your jurisdiction, the storage capacity of the datastore cluster, and how long clients may legitimately need to query historical entities.

The defaults shipped with the deployment tooling (`30m` TTL on RID running hourly, `2688h` ≈ 112-day TTL on SCD running nightly when enabled) are starting points, not recommendations. Validate production values against the criteria above.

Additionally, it is important to note that in a DSS pool, all DSS instances should be running the cleanup, even though it removes entities belonging to all instances.
First, out of governance concerns, as this avoids an undesirable asymmetry between DSS instances. And second, doing so brings redundancy across the pool, in case one of the instances fails to properly perform the cleanup.

## Performance impact

Each entity type (SCD operational intents, SCD subscriptions, RID ISAs, RID subscriptions) is evicted with a single statement. Entity types are deleted independently, so contention or failure on one does not affect the others. There is no risk of data inconsistency in this case - the cleanup may simply be retried.

To mitigate contention:

- Run the cleanup during low-intensity periods (e.g. at night).
- Clean up iteratively, starting with a lower TTL and progressively increasing it.
- Use `--rid_limit`/`--scd_limit` to bound the size of each deletion if a batch still causes performance issues.
- Within a DSS pool, offset the cleanup schedules of the different instances to avoid having them running concurrently.

## How to run the cleanup

The `evict` subcommand of `db-manager` is a CLI tool that lists and deletes expired entities in the DSS store.
It supports cleanup of the following entity types: SCD operational intents, SCD subscriptions, RID identification service areas (ISAs), and RID subscriptions.
Expiration of entities is preferably determined through their end times. In the unusual event that an end time is not available, the last update time is used instead.
All cleanups are global operations that remove expired entities across all DSS instances in the pool, regardless of which instance created them.

The usage of this tool is preferably automated by scheduling regular cleanups via a Kubernetes `CronJob`. But if necessary, it can be invoked manually.

### Automated regular cleanup via Kubernetes `CronJob`

The supported Terraform-based deployments include the RID and SCD eviction cron jobs. Those are shipped with safe defaults but should probably be tuned to your DSS pool:

- RID: enabled by default, schedule: `<min> * * * *`, TTL of 30 minutes;
- SCD: disabled by default, schedule `<min> 2 * * *`, TTL of 2688 hours (2 x 56 days).

Where `<min>` is a minute (`0`–`59`) deterministically derived from the DSS instance's `locality` when `schedule` is left empty (`""`).
As noted in [Why and when to run the cleanup](#why-and-when-to-run-the-cleanup) all DSS instances in a pool should run the cleanup: this default offsets the schedules across instances so they are unlikely to run concurrently.

When deploying via Terraform modules, the parameters are configurable with module variables starting with "evict_".
See [TFVARS.gen.md](https://github.com/interuss/dss/blob/master/deploy/infrastructure/modules/terraform-aws-dss/TFVARS.gen.md) for details.

### Manual ad-hoc cleanup via CLI `db-manager` tool

!!! warning
    Using this tool incorrectly may result in **loss of data**.
    Notably, ensure the clock on the system this tool is run on is reasonably accurate: the expiration of entities is determined based on this clock, and not the server's clock.

Extracted from `db-manager evict --help`:

```
List and evict expired entities

Usage:
  db-manager evict [flags]

Flags:
      --delete             set this flag to true to delete the expired entities
  -h, --help               help for evict
      --rid_isa            set this flag to true to check for expired RID ISAs (default true)
      --rid_limit int      maximum number of RID entities deleted, defaults to unlimited
      --rid_sub            set this flag to true to check for expired RID subscriptions (default true)
      --rid_ttl duration   time-to-live duration used for determining RID entries expiration, defaults to 30 minutes (default 30m0s)
      --scd_limit int      maximum number of SCD entities deleted, defaults to unlimited
      --scd_oir            set this flag to true to check for expired SCD operational intents (default true)
      --scd_sub            set this flag to true to check for expired SCD subscriptions (default true)
      --scd_ttl duration   time-to-live duration used for determining SCD entries expiration, defaults to 2*56 days (default 2688h0m0s)
      --timeout duration   Timeout for the command (default 5m0s)

Global Flags:
      --datastore_application_name string   application name for tagging the connection to the database (default "dss")
      --datastore_host string               database host to connect to
      --datastore_max_conn_idle_secs int    maximum amount of time in seconds a connection may be idle, default is 30 seconds (default 30)
      --datastore_max_open_conns int        maximum number of open connections to the database, default is 4 (default 4)
      --datastore_max_retries int           maximum number of attempts to retry a query in case of contention, default is 100 (default 100)
      --datastore_port int                  database port to connect to (default 26257)
      --datastore_ssl_dir string            directory to ssl certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key
      --datastore_ssl_mode string           database sslmode (default "disable")
      --datastore_user string               database user to authenticate as (default "root")
```

Additional information:

- By default, expired entities are only listed - `--delete` is required to actually remove them.
- A preview run (without `--delete`) and a later `--delete` run are two separate invocations, not one atomic operation. Entities can expire, be created or be removed between the preview and the delete, and the delete run re-evaluates the threshold at its own execution time. With `--rid_limit`/`--scd_limit`, the delete run removes only part of the entities listed by the preview. Keep the gap between the two runs short and re-review if in doubt.
- `--rid_limit`/`--scd_limit` can only be used together with `--delete` and bound the number of entities deleted by a run, per entity type (e.g. `--scd_limit=1000` deletes up to 1000 operational intents and up to 1000 SCD subscriptions). Which expired entities are deleted first is unspecified. Run the command again to delete the remaining ones. A run without `--delete` always lists all expired entities.
- `--locality` is deprecated and has no effect: RID cleanup considers the expired ISAs and subscriptions of all the DSS instances of the pool.
- `--rid_ttl` and `--scd_ttl` accept durations formatted as [Go `time.Duration` strings](https://pkg.go.dev/time#ParseDuration), e.g. `24h`.
- `--timeout` accepts the same duration format and bounds the total execution time of the command.
- The datastore connection flags match those of the `core-service` command.
