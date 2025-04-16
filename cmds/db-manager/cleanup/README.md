# DB Cleanup

## evict
CLI tool that lists and deletes expired entities in the DSS datastore.
At the time of writing this README, the entities supported by this tool are:
- SCD operational intents;
- SCD subscriptions.

The usage of this tool is potentially dangerous: inputting wrong parameters may result in loss of data.
As such it is strongly recommended to always review and validate the list of entities identified as expired, and to
ensure that a backup of the data is available, if appropriate, before deleting anything using the `--delete` flag.

### Performance impact
The current implementation of this tool might have a performance impact due notably to lock contention if the number of
entities to be removed is high. With the system under heavy load it might even fail to remove them. That is due to the
fact that the expired entities are all identified and removed within a single transaction: with concurrent competing
transactions succeeding faster, there might be enough failures so that the tool fails. There is no risk of data
inconsistency and the cleanup may just be tried again in that case.

To avoid this issue:
- perform the cleanup during a low intensity period (e.g. at night);
- iteratively cleanup the entities by starting with a lower TTL and progressively making it higher.

If this becomes enough of an issue in the future it could be considered implementing batching of removals.

### Usage
Extract from running `db-manager evict --help`:
```
List and evict expired entities

Usage:
  db-manager evict [flags]

Flags:
      --delete         set this flag to true to delete the expired entities
  -h, --help           help for evict
      --scd_oir        set this flag to true to list expired SCD operational intents (default true)
      --scd_sub        set this flag to true to list expired SCD subscriptions (default true)
      --ttl duration   time-to-live duration used for determining expiration, defaults to 2*56 days which should be a safe value in most cases (default 2688h0m0s)
```

For global flags, see [datastore flags.go](../../../pkg/datastore/flags/).

Do note:
- by default expired entities are only listed, not deleted, the flag `--delete` is required for deleting entities;
- expiration of entities is preferably determined through their end times, however when they do not have end times, the last update times are used;
- the flag `--ttl` accepts durations formatted as [Go `time.Duration` strings](https://pkg.go.dev/time#ParseDuration), e.g. `24h`;
- the datastore connection flags are the same as [the `core-service` command](../../core-service/README.md).

### Examples
The following examples assume a running DSS deployed locally through [the `run_locally.sh` script](../../../build/dev/standalone_instance.md).

#### List all entities older than 1 week
```shell
docker compose -f docker-compose_dss.yaml -p dss_sandbox exec local-dss-core-service db-manager evict \
 --datastore_host=local-dss-crdb --ttl=168h
```

#### List operational intents older than 1 week
```shell
docker compose -f docker-compose_dss.yaml -p dss_sandbox exec local-dss-core-service db-manager evict \
 --datastore_host=local-dss-crdb --ttl=168h --scd_oir=true --scd_sub=false
```

#### Delete all entities older than 30 days
```shell
docker compose -f docker-compose_dss.yaml -p dss_sandbox exec local-dss-core-service db-manager evict \
 --datastore_host=local-dss-crdb --ttl=720h --delete
```
