# Pooling for deployment

Before [services](../services/index.md) can be deployed to the
[infrastructure](../infrastructure/index.md) of a DSS instance, the pool that the DSS
instance will create or join must be defined.  This section describes how to
define that pooling configuration.  See
[pooling background documentation](../../background/pooling.md) for a
conceptual overview.

## Data stores

The pooling procedure differs depending on data store used by the DSS
instance:

- [CockroachDB pooling](./crdb.md)
- [YugabyteDB pooling](./yugabyte.md)
