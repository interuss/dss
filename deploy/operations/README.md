# Operations

This folder contains the instructions and related material used to operate a DSS. It is responsible to provide diagnostic capabilities and utilities to operate the DSS instance, such as certificates management. 

Currently, the operations scripts are located inside [build](../../build) and if using the [infrastructure layer](../infrastructure), helpers are generated in the workspace directory by terraform after deployment.

As a complete example, the configuration files [used by the CI job](../../.github/workflows/dss-deploy.yml) of the [infrastructure](../infrastructure) and [servics](../services) layers are located in [ci](./ci).

## Pooling procedure

### Creating a new pool

See [Creating a new pool](../../build/pooling.md#creating-a-new-pool)

### Establishing a pool with first instance 

See [Establishing a pool with first instance](../../build/pooling.md#establishing-a-pool-with-first-instance)

### Joining an existing pool with new instance

See [Joining an existing pool with new instance](../../build/pooling.md#joining-an-existing-pool-with-new-instance)

### Leaving a pool

See [Leaving a pool](../../build/pooling.md#leaving-a-pool)

## Troubleshooting

See [Troubleshooting](../../build/README.md#troubleshooting)
