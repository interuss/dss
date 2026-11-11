# CockroachDB Pool Survivability

This document describes how to meet the survivability and high-availability objectives of a DSS pool running on CockroachDB (as described in the [Survivability section of the Architecture Overview](./architecture.md#Survivability)).

For example, in a standard pool of 3 DSS providers running 2 CockroachDB nodes each (6 nodes total):
* If we do not restrict where data replicas are placed, CockroachDB might place two of the three replicas of a range on nodes belonging to a single DSS provider.
* If that DSS provider suffers an outage or goes down for maintenance, we would lose two replicas at once. This breaks the Raft quorum (since only 1 of 3 replicas remains active), causing that range to become unavailable.
* To prevent this and meet survivability objectives, we must configure the CockroachDB pool to place at least one replica per DSS provider. This guarantees that even if a full DSS provider goes down, 2 out of 3 replicas remain active on the other ones, maintaining quorum and ensuring uninterrupted DSS operations.

To achieve this, we:
1. Start each CockroachDB node with a specific locality embedding its DSS provider and node identifier.
2. Configure replication zones using the CockroachDB `ALTER RANGE` SQL command to enforce placement constraints across the pool.

---

## 1. Setting Node Locality

When starting each CockroachDB node, you must configure its `--locality` flag to specify which DSS provider and node it represents. For the purpose of ensuring survivability of the cluster, it is important that all DSS providers align of the structure of their locality settings.
We recommend that the locality use first at least the DSS provider key (i.e. `dss`), and optionally second the `node` key.
Because CockroachDB will distribute range replicas using values of locality keys, the order of the keys defining the hierarchy to do so, if there is some existing locality set preventing this to be achieved, it is important that:
- all DSS providers agree on a specific key whose value is specific to them (e.g. `dss=dss1`);
- and that the number of total replicas set covers at least the total number of combination of the DSS provider key and the ones before.

Example: with three DSS providers each with two nodes define the following localities:
- `region=east,dss=dss1,node=dss1_node1`
- `region=west,dss=dss1,node=dss1_node2`
- `region=east,dss=dss2,node=dss2_node1`
- `region=west,dss=dss2,node=dss2_node2`
- `region=east,dss=dss3,node=dss3_node3`
- `region=west,dss=dss3,node=dss3_node3`
Then, the total number of replicas should be set to at least 6 (2 `region` * 3 `dss`).

If, for some reason specific to your deployment, this is not possible to achieve, you will need to configure specific range constraints on your deployment to ensure survivability. This is however not within the scope of this documentation.

!!! danger "Ordering Constraint"
    The ordering of the `--locality` flag keys must be exactly the same across all CockroachDB nodes in the cluster (e.g., `dss` first, then `node`). Mixing the order (e.g., `node` then `dss` on some nodes) will cause CockroachDB to treat them as incompatible locality hierarchies and fail to apply constraints correctly.

### Flag Format

```shell
--locality=dss=<dss_id>,node=<node_id>
# or
--locality=dss=<dss_id>
```

Where:
* `<dss_id>` is a unique identifier for the DSS provider (e.g., `dss1`, `dss2`, `dss3`).
* `<node_id>` is a unique identifier for the node within that DSS provider (e.g., `node-0`, `node-1`).

### Example Configuration for a pool of 3 DSS providers

* **DSS 1 (dss1)**:
  * Node 0: `--locality=dss=dss1,node=node-0`
  * Node 1: `--locality=dss=dss1,node=node-1`
* **DSS 2 (dss2)**:
  * Node 0: `--locality=dss=dss2,node=node-0`
  * Node 1: `--locality=dss=dss2,node=node-1`
* **DSS 3 (dss3)**:
  * Node 0: `--locality=dss=dss3,node=node-0`
  * Node 1: `--locality=dss=dss3,node=node-1`


## 2. Configuring Replication Constraints (`ALTER RANGE`)

By default, CockroachDB automatically distributes replicas to optimize resource usage and load. To enforce the "one replica per DSS" survivability rule, you must manually define replication zone constraints using the `ALTER RANGE default CONFIGURE ZONE` SQL command.

The `default` range is the cluster-wide catch-all. Any database or table created within the DSS (including the RID and SCD tables) that does not have its own specific zone configuration will inherit these default placement constraints.

### The `ALTER RANGE` SQL Statement

To configure a 3-replica cluster where exactly one replica lives on each of the three DSSs, assuming locality is set on each node as described above:

```sql
ALTER RANGE default CONFIGURE ZONE USING
  num_replicas = 3,
  constraints = '{"+dss=dss1": 1, "+dss=dss2": 1, "+dss=dss3": 1}';
```

### Explanation of the Parameters:
* `num_replicas = 3`: Tells CockroachDB to keep 3 copies of each range.
* `constraints`: A JSON object specifying per-replica constraints. Consult the [CockroachDB documentation about replication controls](https://docs.cockroachlabs.com/docs/v24.1/configure-replication-zones) for handling more complex use-cases.
    * `"+dss=dss1": 1` tells CockroachDB that exactly one replica must be located on nodes matching the locality `dss=dss1`.

Do note that outside of the default case described in this documentation, you will need to adjust the above SQL query by configuring different ranges and/or setting specific constraints. This is however not within the scope of this documentation.
