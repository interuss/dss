# DSS Pooling

## Introduction

The DSS is designed to be deployed in a federated manner where multiple
organizations each host a DSS instance, and all of those instances interoperate.
Specifically, if a change is made on one DSS instance, that change may be read
from a different DSS instance.  A set of interoperable DSS instances is called a
"pool", and the purpose of this document is to describe how to form and maintain
a DSS pool.

It is expected that there will be exactly one production DSS pool for any given
DSS region, and that a DSS region will generally match aviation jurisdictional
boundaries (usually national boundaries).  A given DSS region (e.g.,
Switzerland) will likely have one pool for production operations, and an
additional pool for partner qualification and testing (per, e.g.,
F3411-19 A2.6.2).

### Terminology notes

CockroachDB (CRDB) establishes a distributed data store called a "cluster".
This cluster stores the DSS Airspace Representation (DAR) in multiple SQL
databases within that cluster.  This cluster is composed of many CRDB nodes,
potentially hosted by multiple organizations.

Kubernetes manages a set of services in a "cluster".  This is an entirely
different thing from the CRDB data store, and this type of cluster is what the
deployment instructions refer to.  A Kubernetes cluster contains one or more
node pools: collections of machines available to run jobs.  This node pool is an
entirely different thing from a DSS pool.

## Objective

A pool of InterUSS-compatible DSS instances is established when all of the
following requirements are met:

1. Each CockroachDB node is addressable by every other CockroachDB node
1. Each CockroachDB node is discoverable
1. Each CockroachDB node accepts the certificates of every other node
1. The CockroachDB cluster is initialized

The procedures in this document are intended to achieve all the objectives
listed above, but these procedures are not the only ways to achieve the
objectives.

### "Each CockroachDB node is addressable by every other CockroachDB node"

Every CRDB node must have its own externally-accessible hostname (e.g.,
1.db.dss-prod.example.com), or its own hostname:port combination (e.g.,
db.dss-prod.example.com:26258).  The port on which CRDB communicates must be
open (default CRDB port is 26257).

This requirement may be verified by conducting a standard TLS diagnostic
(like [this one](https://www.wormly.com/test_ssl)) on the hostname:port
for each CRDB node (e.g., 0.db.dss.example.com:26257).  The "Trust"
characteristic will not pass because the certificate is issued by
"Cockroach CA" which is not a generally-trusted root CA, but we
explicitly enable trust by manually exchanging the trusted CA public keys
in ca.crt (see "Each CockroachDB node accepts the certificates of every other
node" below).  However, all other checks should generally pass.

### "Each CockroachDB node is discoverable"

When a CRDB node is brought online, it must know how to connect to the
existing network of nodes.  This is accomplished by providing an explicit
list of nodes to contact.  Each node contacted will provide a list of
nodes it is connected to in the network ("gossip"), so not every node
must be present in the explicit list.  The explicit list should contain,
at a minimum, all nodes within the DSS instance, and at least one node
from each other DSS instance.  Standard practice should generally include
two nodes from each other DSS instance in case one node is down for
maintenance.

### "Each CockroachDB node accepts the certificates of every other node"

CockroachDB uses TLS to secure connections, and TLS includes a mechanism to
ensure the identity of the server being contacted.  This mechanism requires a
trusted root Certificate Authority (CA) to sign a certificate containing the
public key of a particular server, so a client connecting to that server can
verify that the certificate (containing the public key) is endorsed by the root
CA as being genuine.  CRDB certificates require a claim that standard web CAs
will not sign, so instead each USS acts as their own root CA.  When USS 1
is presented with certificates signed by USS 2's CA, USS 1 must know that it
can trust that certificate.  We accomplish this by exchanging all USSs' CA
public keys out-of-band in ca.crt, and specifying that certificates signed by
any of the public keys in ca.crt should be accepted when considering the
validity of certificates presented to establish a TLS connection between nodes.

The private CA key make-certs.py generates is stored in ca_key_dir, and its
corresponding public key is stored in ca_certs_dir (ca.crt).  The private CA key
is used to generate all node certificates and client certificates.  Once a pool
is established, a USS avoids regenerating this CA keypair, and use the existing
ones by default. This default behavior can be modified by setting
--overwrite-ca-cert flag to true.
If a USS generates a new CA keypair, the new public key must be added to the pool's
combined ca.crt, and all USSs in the pool must adopt the new combined ca.crt before
any nodes using certificates generated by the new CA private key will be accepted by the pool.

### "The CockroachDB cluster is initialized"

A CRDB cluster of databases is like the Ship of Theseus: it is composed of
many nodes which may all be replaced, one by one, so that a given CRDB cluster
eventually contains none of its original nodes.  Unlike the Ship of Theseus,
however, a cluster is clearly identified by its cluster ID (e.g.,
b2537de3-166f-42c4-aae1-742e094b8349) -- if the cluster ID is the same, it is
the same cluster (and vice versa).  Once for the entire lifetime of the CRDB
cluster, it is created and given its cluster ID with the `cockroach init`
command.  If this command is run on a node that is part of a cluster that has
already been initialized, it will fail.  The Kubernetes configuration in this
repository uses the `VAR_SHOULD_INIT` parameter to control whether this command
is executed during deployment.

## Additional requirements

These requirements must be met by every DSS instance joining an
InterUSS-compatible pool.  The deployment instructions produce a system that
complies with all these requirements, so this section may be ignored if
following those instructions.

- All CockroachDB nodes must be run in secure mode, by supplying the
  `--certs-dir` and `--ca-key` flags.
  - Do not specify `--insecure`
- The ordering of the `--locality` flag keys must be the same across all
  CockroachDB nodes in the cluster.
- All DSS instances in the same cluster must point their ntpd at the same NTP
  Servers.
  [CockroachDB recommends](https://www.cockroachlabs.com/docs/stable/recommended-production-settings.html#considerations)
  using
  [Google's Public NTP](https://developers.google.com/time/) when running in a
  multi-cloud environment.

## Creating a new pool
Although all DSS instances are equal peers, one DSS instance must be chosen to
create the pool initially.  After the pool is established, one additional DSS
instance can join it.  After that joining process is complete, it can be
repeated any number of times to add additional DSS instances, though 7 is the
maximum recommended number of DSS instances for performance reasons.  The
following diagram illustrates the pooling process for the first two instances:

![DSS pooling for first two participants](../assets/generated/create_pool_2.png)

The nth instance joins in almost the same way as the second instance; the
diagram below illustrates action dependencies between the existing and joining
DSS instances to allow the new instance to join the existing pool:

![DSS pooling for nth participants](../assets/generated/create_pool_n.png)

### Establishing a pool with first instance
The USS owning the first DSS instance should follow
[the deployment instructions](README.md).  They are not joining any existing
cluster, and specifically `VAR_SHOULD_INIT` should be set `true` to initialize
the CRDB cluster.  Upon deployment completion, the following should be run against the DSS instance to verify functionality:

 - The [prober test](https://github.com/interuss/monitoring/blob/main/monitoring/prober/README.md)
 - The [USS qualifier](https://github.com/interuss/monitoring/tree/main/monitoring/uss_qualifier), using the [DSS Probing](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/configurations/dev/dss_probing.yaml) configuration


### Joining an existing pool with new instance
A USS wishing to join an existing pool (of perhaps just one instance following
the prior section) should follow [the deployment instructions](README.md).  They
will be joining an existing cluster, and they will need to request the ca.crt
that the pool is currently using (any one member of the pool may provide it).
The joining USS will also need a list of node addresses to which they should
connect; ideally this will include at least 2 nodes from each existing DSS
instance.

As soon as the joining USS has created a new, combined ca.crt by joining the
existing ca.crt provided by the pool, the joining USS must provide that new
ca.crt to every existing DSS instance in the pool, and all nodes of all
instances must adopt the new, combined ca.crt before the joining USS can bring
their DSS instance online (with `tk apply`).

Once all participants in the existing pool have confirmed that the new ca.crt
has been adopted by all of their nodes, the joining USS brings their system
online with `tk apply`.  The pool should then be re-verified for functionality
by running the prober test on each DSS instance, and the
[interoperability test scenario](https://github.com/interuss/monitoring/blob/main/monitoring/uss_qualifier/scenarios/astm/netrid/v19/dss_interoperability.md)
on the full pool (including the newly-added instance).

Finally, the joining USS should provide its node addresses to all other
participants in the pool, and each other participant should add those addresses
to the list of CRDB nodes their CRDB nodes will attempt to contact upon restart.

## Leaving a pool

In an event that requires removing CockroachDB nodes we need to properly and
safely decommission to reduce risks of outages.

It is never a good idea to take down more than half the number of nodes
available in your cluster as doing so would break quorum. If you need to take
down that many nodes please do it in smaller steps.

Note: If you are removing a specific node in a Statefulset, please know that
Kubernetes does not support removal of specific node; it automatically
re-creates the node if you delete it with `kubectl delete pod`.  You will need
to scale down the Statefulset and that removes the last node first (ex:
`cockroachdb-n` where `n` is the `size of statefulset - 1`, `n` starts at 0)

1. Check if all nodes are healthy and there are no
   under-replicated/unavailable ranges:

   `kubectl exec -it cockroachdb-0 -- cockroach node status --ranges --certs-dir=cockroach-certs/`

    1. If there are under-replicated ranges changes are it is because of a node
       failure. If all nodes are healthy then it should auto recover.

    1. If there are unhealthy nodes please investigate and fix them so that the
       ranges can return to a healthy state

1. Identify the node id we intend to decommission from the previous commands
   then decommission them. The following command assumes that `cockroachdb-0` is
   not targeted for decommission otherwise select a different instance to
   connect to:

   `kubectl exec -it cockroachdb-0 -- cockroach node decommission <node id 1> [<node id 2> ...] --certs-dir=cockroach-certs/`

1. If the command executes successfully all targeted nodes should not host any
   ranges. Repeat step one to verify

    a. In the event of a hung decommission please recommission the nodes and
    repeat the above step with smaller number of nodes to decommission:

    `kubectl exec -it cockroachdb-0 -- cockroach node recommission <node id 1> [<node id 2> ...] --certs-dir=cockroach-certs/`

1. Power down the pods or delete the Statefulset, whichever is applicable

    a. Again, Statefulsets does not support deleting specific pods, as it will
       restart it immediately you will need to scale down understanding that it
       will remove node `cockroachdb-n` first; where `n` is the
       `size of statefulset - 1`.

       To proceed: `kubectl scale statefulset cockroachdb --replicas=<X>`

    b. To remove the entire Statefulset:
    `kubectl delete statefulset cockroachdb`
