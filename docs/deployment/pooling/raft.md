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

The day-2 operations on the pool (adding or removing nodes, replacing a
node) are described in [pooling operations](../../operations/pooling.md);
the nodes are restarted as described in
[modifying a deployment](../../operations/modifying-deployment.md#restarting-nodes).

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
   number (0-based, the order of the pods of the deployment) and `<TYPE>` is
   replaced by `raft`:

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
    must be [restarted](../../operations/modifying-deployment.md#restarting-nodes).

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

   and [restarts its nodes](../../operations/modifying-deployment.md#restarting-nodes):
   the nodes read the CA bundle only at startup, so they must be restarted to
   accept the new CA.

1. Every participant should then compare the hash of its pool of CAs
   (displayed by `add-pool-ca` and by
   `./dss-certs.py ... list-pool-ca`) to make sure they work with the same
   coherent set of certificates.

!!! danger
    Do not proceed until every participant of the existing pool has positively
    confirmed that its DSS instance accepts the CA of this DSS instance.

## Forming the pool

### If you are the first participant

This DSS instance creates the pool (it is the first instance, or one of the
instances creating the pool at the same time as other participants).

1. In the deployment configuration of this instance, enable the
   **automatic bootstrap** feature (for example
   `raft_automatic_bootstrap` in the terraform variables, or
   `raft.automaticBootstrap` in the helm chart).

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
   ([inspecting the membership](../../operations/pooling.md#inspecting-the-membership)).

Without the automatic bootstrap, the first node of the pool is bootstrapped
and the other nodes are added manually with the `dssctl raft` commands: see
[pooling operations](../../operations/pooling.md).

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
   time: see [adding a node](../../operations/pooling.md#adding-a-node).
   Once a node is added, the change is known to every member of the pool: no
   other participant has to do anything for the membership.

1. Verify that the new nodes are `member` in every Raft group on this
   instance and on the existing instances
   ([inspecting the membership](../../operations/pooling.md#inspecting-the-membership)).

## Next steps

Proceed to [services deployment](../services/index.md).

Once the pool is formed, its day-2 operations are described in
[pooling operations](../../operations/pooling.md).

!!! note
    The eviction cron jobs that the deployment tools provide for the SQL
    datastores are not deployed yet for the Raft store.
