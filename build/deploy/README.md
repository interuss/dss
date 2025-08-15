# Kubernetes deployment via Tanka

This documentation has been moved to [interuss.github.io/dss](https://interuss.github.io/dss).

## Migrating configurations to new location

The following steps describe how to update your workspace configurations to use the new configuration location.

For tanka only deployments, update imports in your `main.jsonnet` for `dss` and `metadataBase` libraries.
Replace the current paths with:
```
local dss = import '../../../deploy/services/tanka/dss.libsonnet';
local metadataBase = import '../../../deploy/services/tanka/metadata_base.libsonnet';
```
