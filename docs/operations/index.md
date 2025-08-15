# Operations

This folder contains the instructions and related material used to operate a DSS. It is responsible to provide diagnostic capabilities and utilities to operate the DSS instance, such as certificates management.

Currently, the operations scripts are located inside [build](../build.md) and if using the [infrastructure layer](../infrastructure/index.md), helpers are generated in the workspace directory by terraform after deployment.

As a complete example, the configuration files [used by the CI job](https://github.com/interuss/dss/blob/master/.github/workflows/dss-deploy.yml) of the [infrastructure](../infrastructure/index.md) and [services](../services/index.md) layers are located in [ci](ci/index.md).

## Pooling procedure

### Creating a new pool

See [Creating a new pool](pooling.md#creating-a-new-pool)

### Establishing a pool with first instance

See [Establishing a pool with first instance](pooling.md#establishing-a-pool-with-first-instances)

### Joining an existing pool with new instance

See [Joining an existing pool with new instance](pooling.md#joining-an-existing-pool-with-new-instance)

### Leaving a pool

See [Leaving a pool](pooling.md#leaving-a-pool)

## Monitoring

See [Monitoring](monitoring.md)

## Troubleshooting

See [Troubleshooting](troubleshooting.md)
