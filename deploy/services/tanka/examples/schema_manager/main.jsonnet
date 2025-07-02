# The path used in imports below must be updated to point to /deploy/services/tanka/

local schemaManager = import '../schema-manager.libsonnet';
local metadataBase = import '../metadata_base.libsonnet';

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
    image: 'VAR_DOCKER_IMAGE_NAME',
    desired_rid_db_version: '4.0.0',
    desired_scd_db_version: '3.2.0',
    desired_aux_db_version: '1.0.0',
  },
};

schemaManager.all(metadata)
