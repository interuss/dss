# The path used in imports below must be updated to point to /deploy/services/tanka/

local schemaManager = import '../schema-manager.libsonnet';
local metadataBase = import '../metadata_base.libsonnet';

// All VAR_* values below must be replaced with appropriate values; see
// dss/build/README.md for more information.

local rid_db_version = importstr "../../../../../build/db_schemas/rid.version"
local scd_db_version = importstr "../../../../../build/db_schemas/scd.version"

local metadata = metadataBase {
  namespace: 'VAR_NAMESPACE',
  clusterName: 'VAR_CLUSTER_CONTEXT',
  enableScd: false, // <-- This boolean value is VAR_ENABLE_SCD
  cockroach+: {
    shouldInit: true, // <-- This boolean value is VAR_SHOULD_INIT
    JoinExisting: [], // <-- This must be set to empty
  },
  schema_manager+: {
    enable: true, // <-- this boolean value is VAR_ENABLE_SCHEMA_MANAGER
    image: 'VAR_DOCKER_IMAGE_NAME',
    desired_rid_db_version: rid_db_version,
    desired_scd_db_version: scd_db_version,
    desired_aux_db_version: '1.0.0',
  },
};

schemaManager.all(metadata)
