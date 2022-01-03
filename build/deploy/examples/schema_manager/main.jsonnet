local schemaManager = import '../../../deploy/schema-manager.libsonnet';
local metadataBase = import '../../../deploy/metadata_base.libsonnet';

// All VAR_* values below must be replaced with appropriate values; see
// dss/build/README.md for more information.

local metadata = metadataBase {
  namespace: 'VAR_NAMESPACE',
  clusterName: 'VAR_CLUSTER_CONTEXT',
  enableScd: false, // <-- This boolean value is VAR_ENABLE_SCD
  cockroach+: {
    shouldInit: true, // <-- This boolean value is VAR_SHOULD_INIT
    JoinExisting: [], // <-- This must be set to empty
  },
  schema_manager+: {
    image: 'VAR_SCHEMA_MANAGER_IMAGE_NAME',
    desired_rid_db_version: '4.0.0',
    desired_scd_db_version: '3.1.0',
  },
};

schemaManager.all(metadata)
