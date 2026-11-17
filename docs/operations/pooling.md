# Pooling operations

Pooling of a DSS instance (having it join an existing pool) is accomplished in the normal course of [deployment](../deployment/index.md).

An existing DSS instance cannot be moved to a different pool.  Instead, the existing DSS instance must be decommissioned and a new DSS instance deployed to the different pool.

Depooling a DSS instance (removing it from the pool) is covered in [decommissioning](../decommissioning/index.md).

## Raft store

!!! warning "Experimental"
    The Raft store is experimental: its implementation is still in progress and
    the procedures in this section may change.

This section describes how to manage the membership of a pool running the
[Raft store](../deployment/pooling/raft.md).  Restarting the nodes, for
instance after updating their certificates, is described in
[modifying a deployment](modifying-deployment.md#restarting-nodes).

A DSS node running the Raft store is a member of three Raft groups (one per
service `aux`, `rid` and `scd`)  each with its own membership.  Adding a node
to the pool means adding it to the three groups; the same applies to removal.
The membership is managed with the `dssctl raft` commands, which talk to the
admin API of a node over the Unix socket it serves at `--admin_socket`
(default `/run/dss/admin.sock`).  Any node of the pool can be used: a
participant may manage the pool through any node it can reach.

### Running dssctl

`dssctl` is in the DSS image, inside every core-service pod:

```shell
kubectl --context $CLUSTER_CONTEXT exec --namespace $NAMESPACE -it $POD -- dssctl raft <command>
```

where `$POD` is the name of a core-service pod of any node of the pool
(`<release>-core-service-0` with the helm chart, e.g. `dss-core-service-0`, or
`core-service-0` with the tanka files).

Common options:

| Option | Description |
| :--- | :--- |
| `--groups` | Comma-separated list of the Raft groups to act on (`aux`, `rid`, `scd`); all of them if not set. |
| `--admin_socket` | Path of the admin socket of the node, matching its `--admin_socket`; default `/run/dss/admin.sock`. |
| `--admin_timeout` | Timeout of each request to the admin API; default 1m. |

A command is executed **independently for each selected group**: it may
succeed for some groups and fail for others.  When that happens, rerun the
command with `--groups` limited to the failed groups.

### Inspecting the membership

`dssctl raft node` describes the node serving the admin API in each of its
groups: its state (`waiting`, `member` or `removed`), its ID and its URL.

```shell
dssctl raft node
# aux: member (ID 2213958870532660157, https://0.raft.dss.example.com:2407)
# rid: member (ID 1263815311569516745, https://0.raft.dss.example.com:2407)
# scd: member (ID 4401297763280950187, https://0.raft.dss.example.com:2407)
```

`dssctl raft list` lists, for each group, the nodes that are or were members
of it, sorted by URL (so that the groups list the nodes in the same order),
with the current leader marked:

```shell
dssctl raft list
# aux:
#   ID                    URL                                       ROLE
#   2213958870532660157   https://0.raft.dss.example.com:2407       leader
#   1263815311569516745   https://1.raft.dss.example.com:2407       voter
#   4401297763280950187   https://2.raft.dss.example.com:2407       voter
#   removed IDs: [5740029381146209443]
```

A node has a different ID in each group: the IDs are generated when the node
starts for the first time with an empty data directory and are stored in it.

### Adding a node

A node is added to the pool one at a time, from any member of the pool.  The
node to add must:

* be **running and waiting**, i.e. started with an empty data directory (a
  node that was never added to the pool, or whose data directory was deleted);
* have its participant's CA in the CA bundle of every member of the pool,
  otherwise the TLS handshake between the members and the node fails
  (see [certificates](../deployment/pooling/raft.md#certificates));
* be reachable by the members at its `--raft_peer_url` (its `ENDPOINT`);
* use a URL that is not already used by another member of the pool.

```shell
dssctl raft add --url https://<ENDPOINT>
# aux: added node 993958870532660157 (https://3.raft.dss.example.com:2407)
# rid: added node 463815311569516745 (https://3.raft.dss.example.com:2407)
# scd: added node 2401297763280950187 (https://3.raft.dss.example.com:2407)
```

The member reads the ID of the node to add from the node itself, on its Raft
listener; the node then learns the configuration and the data of the group
from the leader.  The added node joins directly as a **voter**: as soon as the
command returns, it counts in the quorum of the group.  The command returns
once the change is adopted by the group.

Verify the addition with `dssctl raft list` on the node used and on at least
one other participant's node: the new node should be `member` in every group.

### Removing a node

A node is removed one at a time, from any member of the pool:

```shell
dssctl raft remove --url https://<ENDPOINT>
# aux: removed node 993958870532660157
# rid: removed node 463815311569516745
# scd: removed node 2401297763280950187
```

or, for a single group selected with `--groups`, by its ID in that group:

```shell
dssctl raft remove --groups aux --node_id 993958870532660157
```

Properties of a removal:

* A removed node **stops itself** when the removal is applied and its state
  becomes `removed`.  It does not need to be reachable for the removal to
  succeed.
* The ID of a removed node is **retired permanently**: it can never be added
  back to the group.  A node that is removed and re-added comes back with a
  new ID (see [Replacing a node](#replacing-a-node)).
* A voter cannot be removed if the group would be left with **fewer than
  three voters**.
* Removing a voter changes the quorum immediately, and removing the current
  leader triggers an election.

Verify the removal with `dssctl raft list`: the node is gone from the members
and its ID is in the list of removed IDs.

### Replacing a node

If a node must come back at the **same URL** with a new data directory
(typically after its data directory was lost or corrupted):

1. With four voters or more in each group: remove the node
   (`dssctl raft remove --url https://<ENDPOINT>`), delete the data directory
   of the pod (the `raft-data` persistent volume), restart the pod (it
   generates new IDs and waits) and add it back (`dssctl raft add --url
   https://<ENDPOINT>`).

1. With exactly three voters in a group, the group cannot go below the minimum
   of three voters, so the node must be restarted with a **different**
   `--raft_peer_url` (a new `ENDPOINT`) and added to the groups **before** its
   old identity is removed.

In both cases the node comes back as a new node, with new IDs in every group:
it receives a snapshot from the pool and catches up on the entries since.

### After a membership change

* Verify the membership on at least two participants of the pool: every
  member of a group holds the same membership once the change is adopted.
* If a participant leaves the pool entirely, the remaining participants should
  eventually remove its CA from their pool of CAs
* The certificates must be rotated independently of the membership: see
  [certificates](../deployment/pooling/raft.md#certificates).
