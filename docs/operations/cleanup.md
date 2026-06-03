# Database cleanup

Data will accumulate over time in the database if client USSs do not remove their expired entities, and this can lead to lower performance due to quantity of stale entities.  For this reason, InterUSS recommends periodically cleaning up no-longer-relevant entities if USS clients do not always clean up after themselves.

This page describes how to clean up expired entities from the DSS datastore using the `db-manager evict` command.

## Overview

The `evict` subcommand of `db-manager` is a CLI tool that lists and deletes expired entities in the DSS store. The following entity types are supported:

- SCD operational intents
- SCD subscriptions
- RID identification service areas (ISAs)
- RID subscriptions

By default, the tool only lists expired entities. Deletion is opt-in via the `--delete` flag.

!!! warning
    Using this tool incorrectly may result in **loss of data**. Before deleting anything:

    - Always review and validate the list of entities identified as expired (run the tool without `--delete` first).
    - Ensure a backup of the data is available.
    - Double-check the TTL values passed to `--rid_ttl` and `--scd_ttl`.

Expiration of entities is preferably determined through their end times. In the unusual event that an end time is not available, the last update time is used instead.

## Why and when to run the cleanup

Expired entities (operational intents past their end time, stale subscriptions, ISAs whose lifetime has elapsed) can accumulate over time in the datastore if not cleaned up by clients. While the DSS keeps functioning correctly with some stale rows present, excessive accumulation can impact:

- Storage growth: unbounded storage usage in data stores (e.g., CockroachDB / YugabyteDB).
- Query performance: indexes get larger, range scans on `operational_intents` / `subscriptions` / `identification_service_areas` degrade.

There is no single correct interval or TTL. Reasonable values depend on context and must be defined per DSS pool, taking into account: traffic volume and entity churn, regulatory or contractual data-retention requirements applicable to your jurisdiction, the storage capacity of the datastore cluster, and how long clients may legitimately need to query historical entities.

The defaults shipped with the deployment tooling (`30m` TTL on RID running every 30 min, `2688h` ≈ 56-day TTL on SCD running nightly when enabled) are starting points, not recommendations. Validate production values against the criteria above.


## Performance impact

All expired entities are identified and removed within a single transaction. When the system is under heavy load, lock contention with concurrent transactions may cause the cleanup to fail. There is no risk of data inconsistency in this case - the cleanup may simply be retried.

To mitigate this:

- Run the cleanup during low-intensity periods (e.g. at night).
- Clean up iteratively, starting with a lower TTL and progressively increasing it.
If this becomes a recurring issue, batching removals could be considered as a future improvement.

## Changes in locally

There may be cases where a DSS instance changes its locality. A common scenario is that locality was not previously mandatory, though regular updates may also occur.

In such cases, ensure that a cleanup is performed on the older locality, especially since the automatically deployed cron job (see below) will automatically follow the locality settings of the main service.

## Usage

Extracted from `db-manager evict --help`:

```
List and evict expired entities

Usage:
  db-manager evict [flags]

Flags:
      --delete             set this flag to true to delete the expired entities
  -h, --help               help for evict
      --locality string    self-identification string of this DSS instance
      --rid_isa            set this flag to true to check for expired RID ISAs (default true)
      --rid_sub            set this flag to true to check for expired RID subscriptions (default true)
      --rid_ttl duration   time-to-live duration used for determining RID entries expiration, defaults to 30 minutes (default 30m0s)
      --scd_oir            set this flag to true to check for expired SCD operational intents (default true)
      --scd_sub            set this flag to true to check for expired SCD subscriptions (default true)
      --scd_ttl duration   time-to-live duration used for determining SCD entries expiration, defaults to 2*56 days (default 2688h0m0s)

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

Notes:

- By default, expired entities are only listed - `--delete` is required to actually remove them.
- `--rid_ttl` and `--scd_ttl` accept durations formatted as [Go `time.Duration` strings](https://pkg.go.dev/time#ParseDuration), e.g. `24h`.
- The datastore connection flags match those of the `core-service` command.

## Regular cleanup

Beyond running `db-manager evict` manually, the DSS deployment tooling can schedule the cleanup as a recurring Kubernetes `CronJob`. Three deployment paths expose the same set of evict knobs, but will always run with the  `--delete` flag set.

Shared default: RID cleanup is enabled by default (`*/30 * * * *`, `ttl = 30m`); SCD cleanup is disabled by default (suggested schedule `0 2 * * *`, `ttl = 2688h` - i.e. 2 x 56 days).

The current defaults are structured this way because each DSS instance is intended to manage the cleanup of its own RID objects (with deletion restricted to entities created by that specific instance, based on locality). Conversely, SCD cleanup is a global operation; therefore, DSS operators should coordinate to avoid redundant tasks, or potentially designate a single USS to handle the cleanup process.

### Helm

The `dss` chart includes a `dss-evict` CronJob (see `deploy/services/helm-charts/dss/templates/dss-evict.yaml`), configured under `dss.conf.evict` in `values.yaml`:

```
dss:
    conf:
        evict:
            scd:
                enableCron: false
                schedule: "0 2 * * *"
                ttl: 2688h
                operationalIntents: true
                subscriptions: true
            rid:
                enableCron: true
                schedule: "*/30 * * * *"
                ttl: 30m
                ISAs: true
                subscriptions: true
```

### Tanka

 Configure it under the `evict` key of your environment metadata:

```jsonnet
evict+: {
    scd+: {
        enable_cron: false,
        schedule: "0 2 * * *",
        ttl: "2688h",
        operational_intents: true,
        subscriptions: true,
    },
    rid+: {
        enable_cron: true,
        schedule: "*/30 * * * *",
        ttl: "30m",
        ISAs: true,
        subscriptions: true,
    },
},
```

### Terraform

When deploying via terrafrom modules, the parameters are configurable with module variables:

| Terraform variable              |  Default          |
|---------------------------------|------------------|
| `evict_enable_scd_cron`         | `false`          |
| `evict_scd_schedule`            | `"0 2 * * *"`    |
| `evict_scd_ttl`                 | `"2688h"`        |
| `evict_scd_operational_intents` | `true`           |
| `evict_scd_subscriptions`       | `true`           |
| `evict_enable_rid_cron`         | `true`           |
| `evict_rid_schedule`            | `"*/30 * * * *"` |
| `evict_rid_ttl`                 | `"30m"`          |
| `evict_rid_isas`                | `true`           |
| `evict_rid_subscriptions`       | `true`           |

## Examples

The examples below assume a DSS running locally via the `run_locally.sh` script.

### List all entities older than 1 week

```shell
docker compose -f docker-compose_dss.yaml -p dss_sandbox exec local-dss-core-service db-manager evict \
  --datastore_host=local-dss-crdb --scd_ttl=168h --rid_ttl=168h
```

### List only operational intents older than 1 week

```shell
docker compose -f docker-compose_dss.yaml -p dss_sandbox exec local-dss-core-service db-manager evict \
  --datastore_host=local-dss-crdb --scd_ttl=168h --scd_oir=true --scd_sub=false
```

### Delete all entities older than 30 days

```shell
docker compose -f docker-compose_dss.yaml -p dss_sandbox exec local-dss-core-service db-manager evict \
  --datastore_host=local-dss-crdb --scd_ttl=720h --rid_ttl=720h --delete
```
