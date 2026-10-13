# dssctl

## Introduction

`dssctl` is the admin command line tool for a running DSS node. It currently manages the membership of the Raft groups of a core-service using the `raft` datastore (`--store_type=raft`).

It calls the admin API served by the `core-service` of the DSS node on `--admin_addr`. A DSS node using the `raft` datastore runs one Raft group per service: `aux`, `rid` and `scd`. Each group has its own members, and a node joins or leaves each of them independently. A `dssctl raft` command sends one request per selected group, all groups by default.

```
aux: added node 4 (http://dss4.example.com:9704)
scd: added node 4 (http://dss4.example.com:9604)
Error: rid: 400 Bad Request: node ID 4 was removed from the group and cannot be reused (E:…)
```

A change can therefore succeed for some groups and fail for others. In that case, the command can be retried for the failed groups only with `--groups`.

## Usage

### Global Flags

| Flag                | Default                   | Description                                                                   |
| ------------------- | ------------------------- | ----------------------------------------------------------------------------- |
| `--admin_url`     | `http://localhost:8078` | URL of the admin API of the core-service, matching its `--admin_addr` flag. |
| `--admin_timeout` | `1m`                    | Timeout of each request to the admin API.                                     |
| `--groups`        | `aux,rid,scd`           | Comma-separated Raft groups to act on.                                        |

### Commands

### `dssctl raft list`

Lists the members of the Raft groups, as known by the DSS node serving the admin API:

```
rid: as known by node 2
  ID  URL                     ROLE
  1   http://172.27.1.1:9501  leader
  2   http://172.27.2.1:9502  voter
  3   http://172.27.3.1:9503  voter
  removed IDs: [5]
```

### `dssctl raft add --node_id <ID> --<group>_url <URL>...`

Adds a node to the Raft groups, as a voter. The URL at which the other members reach the node in that group is required for each selected group in `--<group>_raft_peers`.

Procedure:

1. Choose an ID for the new node that is not used by any member and was never removed from the groups (see `dssctl raft list`).
2. Start the new node **before** adding it: as soon as it is added, it counts in the quorum of the group. Start it with:
   - `--raft_node_id=<ID>`
   - `--raft_join`, on its first start only, so that it waits to be added instead of creating a new cluster
   - its own entry in each `--<group>_raft_peers` flag, e.g. `--rid_raft_peers=4=http://dss4.example.com:9504`. It learns the other members from the leader
3. From an existing member, add it to the groups:
   ```bash
   dssctl raft add --node_id 4 \
     --aux_url http://dss4.example.com:9704 \
     --rid_url http://dss4.example.com:9504 \
     --scd_url http://dss4.example.com:9604
   ```
4. Check that the node appears in every group, from the new node itself once it is ready: `dssctl raft list`.

### `dssctl raft remove --node_id <ID>`

Removes a node from the Raft groups, whether it is running or not:

```bash
dssctl raft remove --node_id 4
```

- The ID of a removed node is retired permanently: it can never be added again to the same group. A node that must rejoin, e.g. after losing its data, is added with a new ID.
- A voter cannot be removed if the group would be left with fewer than 3 voters.
- A removed node stops participating in the group. If it was the leader, it steps down and the remaining members elect a new one.
- The command can be sent to any member, including the node being removed.
