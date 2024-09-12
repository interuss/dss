# DB Cleanup

## scd-evict
CLI tool that lists and deletes expired entities in the DSS store.
At the time of writing this README, the entities supported by this tool are:
- SCD operational intents;
- SCD subscriptions.

The usage of this tool is potentially dangerous: inputting wrong parameters may result in loss of data.
As such it is strongly recommended to always review and validate the list of entities identified as expired, and to
ensure that a backup of the data is available before deleting anything using the `-delete` flag

### Usage
Extract from running `db-manager scd-evict --help`:
```
List and evict SCD expired entities

Usage:
  db-manager scd-evict [flags]

Flags:
      --delete         set this flag to true to delete the expired entities
  -h, --help           help for scd-evict
      --op_intents     set this flag to true to list expired operational intents (default true)
      --scd_subs       set this flag to true to list expired SCD subscriptions (default true)
      --ttl duration   time-to-live duration used for determining expiration, defaults to 2*56 days which should be a safe value in most cases (default 2688h0m0s)

Global Flags:
      --cockroach_application_name string   application name for tagging the connection to cockroach (default "dss")
      --cockroach_db_name string            application name for tagging the connection to cockroach (default "dss")
      --cockroach_host string               cockroach host to connect to
      --cockroach_max_retries int           maximum number of attempts to retry a query in case of contention, default is 100 (default 100)
      --cockroach_port int                  cockroach port to connect to (default 26257)
      --cockroach_ssl_dir string            directory to ssl certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key
      --cockroach_ssl_mode string           cockroach sslmode (default "disable")
      --cockroach_user string               cockroach user to authenticate as (default "root")
      --max_conn_idle_secs int              maximum amount of time in seconds a connection may be idle, default is 30 seconds (default 30)
      --max_open_conns int                  maximum number of open connections to the database, default is 4 (default 4)

```

Do note:
- by default expired entities are only listed, not deleted, the flag `-delete` is required for deleting entities;
- expiration of entities is preferably determined through their end times, however when they do not have end times, the last update times are used;
- the flag `-ttl` accepts durations formatted as [Go `time.Duration` strings](https://pkg.go.dev/time#ParseDuration), e.g. `24h`;
- the CockroachDB cluster connection flags are the same than [the `core-service` command](../../core-service/README.md).

### Examples
The following examples assume a running DSS deployed locally through [the `run_locally.sh` script](../../../build/dev/standalone_instance.md).

#### List all entities older than 1 week
```shell
docker compose -f docker-compose_dss.yaml -p dss_sandbox exec local-dss-core-service db-evictor \
 -cockroach_host=local-dss-crdb -ttl=168h
```

#### List operational intents older than 1 week
```shell
docker compose -f docker-compose_dss.yaml -p dss_sandbox exec local-dss-core-service db-evictor \
 -cockroach_host=local-dss-crdb -ttl=168h -op_intents=true -scd_subs=false
```

#### Delete all entities older than 30 days
```shell
docker compose -f docker-compose_dss.yaml -p dss_sandbox exec local-dss-core-service db-evictor \
 -cockroach_host=local-dss-crdb -ttl=720h -delete
```
