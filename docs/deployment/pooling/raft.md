# Pooling configuration with the Raft store

!!! warning "Experimental"
    The Raft store is experimental: its implementation is still in progress and
    the procedures in this page may change.

With the Raft store, there is no external database: each node of the pool
stores the data on its own disk and replicates it directly with the other
nodes of the pool through the Raft consensus protocol.  The nodes of the pool
are the datastore of the pool.

To pool a set of DSS instances with the Raft store:

1. [Prerequisites](#prerequisites): the tools and the running
   [infrastructure](../infrastructure/index.md).
2. [Certificates](#certificates): each organization generates its own
   certificate authority (CA) and its nodes' certificates, and the CAs are
   exchanged so that every member of the pool trusts every other.
3. [Forming the pool](#forming-the-pool): the first participant bootstraps
   the Raft groups with its own nodes; whenever a new node joins the pool, an
   existing node adds it.
4. [Managing the pool](#managing-the-pool): the day-2 operations on the
   pool: adding, removing and replacing nodes.

## Prerequisites

* A running Kubernetes platform deployed as
  [infrastructure](../infrastructure/index.md)
* A recent version of python 3 to run the
  [`dss-certs.py` script](https://github.com/interuss/dss/blob/master/deploy/operations/certificates-management/dss-certs.py),
  in the `deploy/operations/certificates-management` folder of this
  repository.

In the instructions below:

* `$CLUSTER_CONTEXT` is the name of the Kubernetes cluster deployed as
  [infrastructure](../infrastructure/index.md) onto which [services](../services/index.md) will be deployed.
* `$NAMESPACE` is the namespace for this DSS instance
* `ENDPOINT` is the `host:port` pair at which the other participants of the
  pool reach the Raft listener of one node of this instance (e.g.
  `0.raft.dss.example.com:2407`). It must be unique in the pool and be
  reachable by the other participants: the deployment tools expose it with one
  load balancer (or static IP) per node.
* `N` is the number of DSS nodes of this instance.

!!! note
    **All participants of a DSS pool must use the same datastore.**  A pool
    cannot mix nodes running the Raft store and nodes running a SQL datastore.

## Certificates

Each participant operates its own certificate authority (CA) and issues one
certificate per node.  Every node holds the PEM bundle of the CAs of *all* the
participants of the pool (including its own) and accepts a Raft connection
only from a peer whose certificate is signed by one of them.  This bundle
defines the boundaries of the pool: it must be identical on every node of the
pool before the nodes can be added to it.

### Creating the certificates of your organization

1. Initialize the certificates of this DSS instance:

    ```shell
    ./dss-certs.py --name <NAME> --cluster-context $CLUSTER_CONTEXT --namespace $NAMESPACE init
    ```

1. If this DSS instance joins an existing pool, import the bundle of the CAs
   of the pool, provided by any one participant of the pool with
   `./dss-certs.py ... get-pool-ca`, saved in `POOL_CA_FILE`:

    ```shell
    ./dss-certs.py --name <NAME> --cluster-context $CLUSTER_CONTEXT --namespace $NAMESPACE add-pool-ca < POOL_CA_FILE
    ```

1. Generate the certificates of the `N` raft nodes.  `--nodes-public-address`
   is the public address of a node, where `<ID>` is replaced by the node
   number (0-based, the order of the pods of the deployment):

    ```shell
    ./dss-certs.py --name <NAME> --cluster-context $CLUSTER_CONTEXT --namespace $NAMESPACE \
      --nodes-count N --nodes-public-address '<ID>.raft.dss.example.com' generate-raft-nodes
    ```

1. Apply the certificates on the cluster to create the `dss.raft.certs`
   secret.  It contains, for each node of this instance, its certificate
   `node.raft-<ID>.crt` and its private key `node.raft-<ID>.key`, and the
   bundle of the CAs of the pool in `ca.crt`:

    ```shell
    ./dss-certs.py --name <NAME> --cluster-context $CLUSTER_CONTEXT --namespace $NAMESPACE apply
    ```

    Run `apply` again after every modification of the certificates (pool CA
    added or removed, new `--nodes-count`).  The nodes read their certificate
    and the CA bundle only at startup, so after the secret is updated they
    must be [restarted](#restarting-the-nodes).

### Adding the certificates of a new organization to an existing pool

When a DSS instance with its own CA joins a pool, every existing participant
must accept the CA of this DSS instance.

1. Get the CA of this DSS instance with
   `./dss-certs.py --name <NAME> --cluster-context $CLUSTER_CONTEXT --namespace $NAMESPACE get-ca`
   and provide it to each participant of the existing pool.

1. Each existing participant, upon receipt of the CA of this DSS instance,
   imports it and updates its certificates:

    ```shell
    ./dss-certs.py --name <OTHER_NAME> --cluster-context <OTHER_CLUSTER_CONTEXT> --namespace <OTHER_NAMESPACE> add-pool-ca < INSTANCE_CA_FILE
    ./dss-certs.py --name <OTHER_NAME> --cluster-context <OTHER_CLUSTER_CONTEXT> --namespace <OTHER_NAMESPACE> apply
    ```

   and [restarts its nodes](#restarting-the-nodes):
   the nodes read the CA bundle only at startup, so they must be restarted to
   accept the new CA.

1. Every participant should then compare the hash of its pool of CAs
   (displayed by `add-pool-ca` and by
   `./dss-certs.py ... list-pool-ca`) to make sure they work with the same
   coherent set of certificates.

!!! danger
    Do not proceed until every participant of the existing pool has positively
    confirmed that its DSS instance accepts the CA of this DSS instance.

### Restarting the nodes

The data of a node (its identity, its Raft log and its snapshots) is stored on
the persistent volume `raft-data` of its pod, mounted at `/raft_data`.  A
restarted node reloads it and rejoins the pool: it replays its log, or
receives a snapshot from a peer if it has been down too long.

The nodes of a participant are restarted with a single command; the stateful
set restarts them one at a time, in reverse pod order (from the highest node
number to the lowest), waiting for a node to be ready before restarting the
next one:

```shell
kubectl --context $CLUSTER_CONTEXT rollout restart statefulset/core-service --namespace $NAMESPACE
```

(With the helm chart, the statefulset is named `<release>-core-service`, e.g.
`dss-core-service`.)

!!! warning
    Do not restart the nodes of the pool all at once, for instance with all
    the participants restarting their nodes at the same time after a CA
    update: a group that loses the majority of its voters stops serving reads
    and writes.  Coordinate the restarts so that a majority of the voters of
    each group stays up.

Use this to restart the nodes whenever the `dss.raft.certs` secret has been
updated (see [certificates](#certificates)).

If the data directory of a node is lost or corrupted, the node must not be
started with an empty one: it would regenerate its identity and simply wait to
be added to the pool, while the pool still remembers its previous one.  Remove
the old identity from the pool, delete the data directory, restart the node
and add it back to the pool: see [replacing a node](#replacing-a-node).

## Forming the pool

### If you are the first participant

This DSS instance creates the pool (it is the first instance, or one of the
instances creating the pool at the same time as other participants).

1. In the deployment configuration of this instance, enable the
   **automatic bootstrap** feature: `raft_automatic_bootstrap` in the
   terraform variables, `raft.automaticBootstrap` in the helm chart, or
   `automaticBootstrap` in the `raft` section of the tanka `main.jsonnet`
   (set automatically from the terraform variables).

    !!! note
        The automatic bootstrap must be enabled on **one** participant only:
        if several nodes bootstrap at the same time, independent pools are
        created.

1. If other participants create their instances at the same time, list the
   `ENDPOINT` of their nodes in the `automaticBootstrapNodes` setting of this
   instance (`raft_automatic_bootstrap_nodes` in the terraform variables).

1. Deploy the services as described in
   [services deployment](../services/index.md).  The other participants
   deploy normally: their nodes start waiting and are added by this instance
   as soon as they are reachable.

1. The first node of this instance bootstraps the three Raft groups
   (`aux`, `rid` and `scd`) with itself as their only member, then adds the
   other nodes of this instance and the nodes of the other participants, one
   at a time, retrying until all of them are members.

1. Check that the nodes have joined the pool: `dssctl raft node` on one of
   them should report `member` in every Raft group, and `dssctl raft list`
   should list all the nodes of the pool
   ([inspecting the membership](#inspecting-the-membership)).

Alternatively, you can do the bootstrap manually with the `dssctl` CLI (see
[running dssctl](#running-dssctl) for how to run `dssctl` against a node of
the pool):

```shell
dssctl raft bootstrap
dssctl raft add --url https://<ENDPOINT>
```

`bootstrap` is run on the first node: it bootstraps the three Raft groups
with itself as their only member.  `add` is then run once for each other node
of the pool, from any member of the pool.

### If you are joining an existing pool

This DSS instance joins a pool that already exists.  Its nodes start with an
empty data directory and wait to be added to the pool by an existing node.

1. Keep the **automatic bootstrap** feature disabled in the deployment
   configuration of this instance.

1. Complete the [certificates of this instance](#creating-the-certificates-of-your-organization)
   and make sure that the existing participants accept its CA (see
   [adding the certificates of a new organization](#adding-the-certificates-of-a-new-organization-to-an-existing-pool)).

1. Deploy the services as described in
   [services deployment](../services/index.md).

1. An existing node adds the nodes of this instance to the pool, one at a
   time: see [adding a node](#adding-a-node).
   Once a node is added, the change is known to every member of the pool: no
   other participant has to do anything for the membership.

1. Verify that the new nodes are `member` in every Raft group on this
   instance and on the existing instances
   ([inspecting the membership](#inspecting-the-membership)).

## Managing the pool

This section describes how to manage the membership of a pool.  A DSS node
running the Raft store is a member of three Raft groups (one per service:
`aux`, `rid` and `scd`), each with its own membership.  Adding a node to the
pool means adding it to the three groups; the same applies to removal.  The
membership is managed with the `dssctl raft` commands, which talk to the
admin API of a node over the Unix socket it serves at `--admin_socket`
(default `/run/dss/admin.sock`).  Any node of the pool can be used: a
participant may manage the pool through any node it can reach.

Restarting the nodes, for instance after updating their certificates, is
described in [restarting the nodes](#restarting-the-nodes).

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
  (see [certificates](#certificates));
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
* If a participant leaves the pool entirely, the remaining participants
  should eventually remove its CA from their pool of CAs: see
  [leaving a pool](../../decommissioning/depooling-raft.md).
* The certificates must be rotated independently of the membership: see
  [certificates](#certificates).

## Next steps

Proceed to [services deployment](../services/index.md).
