# Deployment architecture

## Introduction

The expected deployment configuration of a DSS pool supporting a DSS Region is
multiple organizations to each host one DSS instance that is interoperable with
each other organization's DSS instance.  A DSS pool with three participating
organizations (USSs) will have an architecture similar to the diagram below.

_**Note** that the diagrams below shows 2 stateful sets per DSS instance.  Currently, the helm and tanka deployments produce 3 stateful sets per DSS instance.  However, after Issue #481 is resolved, this is expected to be reduced to 2 stateful sets.  With the Raft store, the instance has a single stateful set: the core-service._

### Certificates

This diagram shows how certificates are shared. It applies to CockroachDB, YugabyteDB and Raft store deployments.

![Pool architecture diagram](../assets/generated/pool_architecture_certs.png)

### CockroachDB

![Pool architecture diagram](../assets/generated/pool_architecture.png)

### Yugabyte

Detail on an instance level
![Pool architecture diagram with Yugabyte](../assets/generated/pool_architecture_yugabyte_instance.png)

Top level simplified view, with one replica shown and yugabyte services regrouped in one box.
![Pool architecture diagram with Yugabyte](../assets/generated/pool_architecture_yugabyte.png)

To reduce the number of required public load balancers, we do use an intermediate reverse proxy to expose the ports of Yugabyte master and tserver on a shared public IP per stateful set instance.
Usual Kubernetes load balancers can't assign connection based on ports out of the box, so we use the reverse proxy to dispatch connections on both services depending on the connected port.

### Raft store

!!! warning "Experimental"
    The Raft store is experimental: its implementation is still in progress.

With the Raft store, the datastore is not a separate cluster, the core-service
nodes of the DSS instances are the datastore of the pool.  Each node keeps the
data of the pool (its identity, its Raft log and its snapshots) on its own
persistent volume and replicates it directly with the other nodes of the pool
through the Raft consensus protocol:


![Pool architecture diagram with the Raft store](../assets/generated/pool_architecture_raft.png)

## Survivability

One of the primary design considerations of the DSS is to be very resilient to
failures.

### CockroachDB

This resiliency is obtained primarily from the behavior of the
underlying CockroachDB database technology and how we configure it. To meet these
survivability objectives, see [CockroachDB Pool Survivability](./survivability-crdb.md) for how to configure
the CockroachDB pool with replica placement constraints.

The diagram below shows the result of failures (bringing a node down for
maintenance, or having an entire USS go down) from different starting points,
assuming 3 replicas.

![Survivability diagram](../assets/generated/survivability_3x2.png)

The table below summarizes survivable failures with 3 DSS instances configured according
to the architecture described above.  Each system state is summarized by three
groups (one group per USS) of two nodes per USS.

* 🟩 : Functional node has no recent changes in functionality
* 🟥 : Non-functional node in down USS has no recent changes in functionality
* 🟧 : Non-functional node due to USS upgrade or maintenance has no recent changes in functionality
* 🔴 : Node becomes non-functional due to a USS going down
* 🟠 : Node becomes non-functional due to USS upgrade or maintenance

| Pre-existing conditions  | New failures | Survivable?
| --- | --- | ---
| (🟩 , 🟩 ) (🟩 , 🟩 ) (🟩 , 🟩 ) | (🟩 , 🟩 ) (🟩 , 🟩 ) (🟩 , 🟠 ) | 🟢 Yes
|                                    | (🟩 , 🟩 ) (🟩 , 🟠 ) (🟩 , 🟠 ) | 🔴 No; some ranges may be lost because of [this bug](https://github.com/cockroachdb/cockroach/issues/66159)
|                                    | (🟩 , 🟠 ) (🟩 , 🟠 ) (🟩 , 🟠 ) | 🔴 No; some ranges may be lost
|                                    | (🟩 , 🟩 ) (🟩 , 🟩 ) (🔴 , 🔴 ) | 🟢 Yes
|                                    | (🟩 , 🟩 ) (🔴 , 🔴 ) (🔴 , 🔴 ) | 🔴 No; ranges guaranteed to be lost
| (🟩 , 🟩 ) (🟩 , 🟩 ) (🟩 , 🟧 ) | (🟩 , 🟩 ) (🟩 , 🟠 ) (🟩 , 🟧 ) | 🟢 Yes
|                                    | (🟩 , 🟠 ) (🟩 , 🟠 ) (🟩 , 🟧 ) | 🔴 No; some ranges may be lost because of [this bug](https://github.com/cockroachdb/cockroach/issues/66159)
|                                    | (🟩 , 🟩 ) (🟩 , 🟩 ) (🔴 , 🔴 ) | 🟢 Yes
|                                    | (🟩 , 🟩 ) (🔴 , 🔴 ) (🟩 , 🟧 ) | 🟡 Yes, with 3 replicas
| (🟩 , 🟩 ) (🟩 , 🟧 ) (🟩 , 🟧 ) | (🟩 , 🟠 ) (🟩 , 🟧 ) (🟩 , 🟧 ) | 🟢 Yes
|                                    | (🟩 , 🟩 ) (🟩 , 🟧 ) (🟠 , 🟧 ) | 🟢 Yes
|                                    | (🟩 , 🟩 ) (🟩 , 🟧 ) (🔴 , 🔴 ) | 🟢 Yes
|                                    | (🔴 , 🔴 ) (🟩 , 🟧 ) (🟩 , 🟧 ) | 🟡 Yes, with 3 replicas
| (🟩 , 🟧 ) (🟩 , 🟧 ) (🟩 , 🟧 ) | (🟩 , 🟧 ) (🟩 , 🟧 ) (🟠 , 🟧 ) | 🟡 Yes, with 3 replicas
|                                    | (🟩 , 🟧 ) (🟠 , 🟧 ) (🟠 , 🟧 ) | 🔴 No; ranges guaranteed to be lost
|                                    | (🟠 , 🟧 ) (🟠 , 🟧 ) (🟠 , 🟧 ) | 🔴 No; ranges guaranteed to be lost
|                                    | (🟩 , 🟧 ) (🟩 , 🟧 ) (🔴 , 🔴 ) | 🟡 Yes, with 3 replicas
| (🟩 , 🟩 ) (🟩 , 🟩 ) (🟥 , 🟥 ) | (🟩 , 🟩 ) (🟩 , 🟠 ) (🟥 , 🟥 ) | 🟡 Yes, with 3 replicas
|                                    | (🟩 , 🟠 ) (🟩 , 🟠 ) (🟥 , 🟥 ) | 🔴 No; some ranges may be lost
|                                    | (🟩 , 🟩 ) (🔴 , 🔴 ) (🟥 , 🟥 ) | 🔴 No; some ranges may be lost

### Raft store

With the Raft store, the resiliency is obtained from the quorum of the Raft
voters instead of the replication of a database: a group of `2N+1` voters
tolerates the loss of `N` of them, and a group that cannot reach a majority of
its voters stops serving reads and writes until one is back.  A restarted or
re-added node receives the data of the pool (it replays its log, or receives
a snapshot from a peer) and rejoins, so no data is lost as long as a majority
of the voters was available.

The table below summarizes survivable failures with 3 DSS instances of 2
nodes each.  Each system state is summarized by three groups (one group per
USS) of two nodes per USS.  Each Raft group then has 6 voters, and its
quorum is 4 of them.  When the quorum is lost, the pool stops serving but no
data is lost: it resumes when a majority of the voters is back.

* 🟩 : Functional node has no recent changes in functionality
* 🟥 : Non-functional node in down USS has no recent changes in functionality
* 🟧 : Non-functional node due to USS upgrade or maintenance has no recent changes in functionality
* 🔴 : Node becomes non-functional due to a USS going down
* 🟠 : Node becomes non-functional due to USS upgrade or maintenance

| Pre-existing conditions  | New failures | Survivable?
| --- | --- | ---
| (🟩 , 🟩 ) (🟩 , 🟩 ) (🟩 , 🟩 ) | (🟩 , 🟩 ) (🟩 , 🟩 ) (🟩 , 🟠 ) | 🟢 Yes
|                                    | (🟩 , 🟩 ) (🟩 , 🟠 ) (🟩 , 🟠 ) | 🟢 Yes
|                                    | (🟩 , 🟩 ) (🔴 , 🔴 ) (🟩 , 🟩 ) | 🟢 Yes
|                                    | (🟩 , 🟩 ) (🟩 , 🟠 ) (🟠 , 🟠 ) | 🔴 No
|                                    | (🟩 , 🟩 ) (🔴 , 🔴 ) (🟩 , 🟠 ) | 🔴 No
|                                    | (🔴 , 🔴 ) (🔴 , 🔴 ) (🟩 , 🟩 ) | 🔴 No
| (🟩 , 🟩 ) (🟩 , 🟩 ) (🟩 , 🟧 ) | (🟩 , 🟩 ) (🟩 , 🟠 ) (🟩 , 🟧 ) | 🟢 Yes
|                                    | (🟩 , 🟩 ) (🔴 , 🔴 ) (🟩 , 🟧 ) | 🔴 No
| (🟩 , 🟩 ) (🟩 , 🟩 ) (🟥 , 🟥 ) | (🟩 , 🟩 ) (🟩 , 🟠 ) (🟥 , 🟥 ) | 🔴 No
|                                    | (🟩 , 🟩 ) (🔴 , 🔴 ) (🟥 , 🟥 ) | 🔴 No

## Sizing

Resources needed are described [in the sizing section](sizing.md)
