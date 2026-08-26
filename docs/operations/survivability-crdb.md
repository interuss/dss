# CockroachDB Pool Survivability

This document describes how to meet the survivability and high-availability objectives of a DSS pool running on CockroachDB (as described in the [Survivability section of the Architecture Overview](../architecture/index.md#Survivability)).

For example, in a standard 3-USS pool running 2 CockroachDB nodes per USS (6 nodes total):
* If we do not restrict where data replicas are placed, CockroachDB might place two of the three replicas of a range on nodes belonging to a single USS.
* If that USS suffers an outage or goes down for maintenance, we would lose two replicas at once. This breaks the Raft quorum (since only 1 of 3 replicas remains active), causing that range to become unavailable.
* To prevent this and meet survivability objectives, we must configure the CockroachDB pool to place at least one replica per USS. This guarantees that even if a full USS goes down, 2 out of 3 replicas remain active on the other USSs, maintaining quorum and ensuring uninterrupted DSS operations.

To achieve this, we:
1. Start each CockroachDB node with a specific locality embedding its USS and node identifier.
2. Configure replication zones using the CockroachDB `ALTER RANGE` SQL command to enforce placement constraints across the pool.

---

## 1. Setting Node Locality

When starting each CockroachDB node, you must configure its `--locality` flag to specify which USS and node it represents. For the purpose of ensuring survivability of the cluster, it is important that all USSes align of the structure of their locality settings.
We recommend that the locality use first at least the `uss` key, and optionally second the `node` key.
Because CockroachDB will distribute range replicas using values of locality keys, the order of the keys defining the hierarchy to do so, if there is some existing locality set preventing this to be achieved, it is important that:
- all USSes define a specific key whose value is specific to them (e.g. `uss=uss1`);
- and that the number of total replicas set is covers at least the total number of combination of keys of to the USS key.

Example: with three USSs each with two nodes define the following localities:
- `region=east,uss=uss1,node=uss1_node1`
- `region=west,uss=uss1,node=uss1_node2`
- `region=east,uss=uss2,node=uss2_node1`
- `region=west,uss=uss2,node=uss2_node2`
- `region=east,uss=uss3,node=uss3_node3`
- `region=west,uss=uss3,node=uss3_node3`
Then, the total number of replicas should be set to at least 6 (2 `region` * 3 `uss`).

If, for some reason specific to your deployment, this is not possible to achieve, you will need to configure specific range constraints on your deployment to ensure survivability. This is however not within the scope of this documentation.

!!! danger "Ordering Constraint"
    The ordering of the `--locality` flag keys must be exactly the same across all CockroachDB nodes in the cluster (e.g., `uss` first, then `node`). Mixing the order (e.g., `node` then `uss` on some nodes) will cause CockroachDB to treat them as incompatible locality hierarchies and fail to apply constraints correctly.

### Flag Format

```shell
--locality=uss=<uss_id>,node=<node_id>
# or
--locality=uss=<uss_id>
```

Where:
* `<uss_id>` is a unique identifier for the USS organization (e.g., `uss1`, `uss2`, `uss3`).
* `<node_id>` is a unique identifier for the node within that USS (e.g., `node-0`, `node-1`).

### Example Configuration for a 3-USS Pool

* **USS 1 (uss1)**:
  * Node 0: `--locality=uss=uss1,node=node-0`
  * Node 1: `--locality=uss=uss1,node=node-1`
* **USS 2 (uss2)**:
  * Node 0: `--locality=uss=uss2,node=node-0`
  * Node 1: `--locality=uss=uss2,node=node-1`
* **USS 3 (uss3)**:
  * Node 0: `--locality=uss=uss3,node=node-0`
  * Node 1: `--locality=uss=uss3,node=node-1`


## 2. Configuring Replication Constraints (`ALTER RANGE`)

By default, CockroachDB automatically distributes replicas to optimize resource usage and load. To enforce the "one replica per USS" survivability rule, you must manually define replication zone constraints using the `ALTER RANGE default CONFIGURE ZONE` SQL command.

The `default` range is the cluster-wide catch-all. Any database or table created within the DSS (including the RID and SCD tables) that does not have its own specific zone configuration will inherit these default placement constraints.

### The `ALTER RANGE` SQL Statement

To configure a 3-replica cluster where exactly one replica lives on each of the three USSs, assuming locality is set on each node as described above:

```sql
ALTER RANGE default CONFIGURE ZONE USING
  num_replicas = 3,
  num_replicas = 3;
```

### Explanation of the Parameters:
* `num_replicas = 3`: Tells CockroachDB to keep 3 copies of each range.

Do note that outside of the default case described in this documentation, you will need to adjust the above SQL query by configuring different ranges and/or setting specific constraints. This is however not within the scope of this documentation.
