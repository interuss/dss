# Kubernetes deployment via Tanka

The documentation and configuration have been moved to [deploy/services](../../deploy/services/tanka).
[Architecture](../../deploy/architecture.md#architecture), [Survivability](../../deploy/architecture.md#survivability) 
and [Sizing](../../deploy/architecture.md#sizing) sections have been moved to [deploy/architecture](../../deploy/architecture.md)

## Migrating configurations to new location

The following steps describe how to update your workspace configurations to use the new configuration location.

For tanka only deployments, update imports in your `main.jsonnet` for `dss` and `metadataBase` libraries.
Replace the current paths with:
```
local dss = import '../../../deploy/services/tanka/dss.libsonnet';
local metadataBase = import '../../../deploy/services/tanka/metadata_base.libsonnet';
```
