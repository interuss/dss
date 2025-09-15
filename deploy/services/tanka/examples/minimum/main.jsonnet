
# The path used in imports below must be updated to point to /deploy/services/tanka/
local dss = import '../dss.libsonnet';
local metadataBase = import '../metadata_base.libsonnet';

// All VAR_* values below must be replaced with appropriate values; see
// dss/build/README.md for more information.

// Crdb versions 
local rid_db_version = importstr "../../db_versions/crdb/rid.version";
local scd_db_version = importstr "../../db_versions/crdb/scd.version";
local aux_db_version = importstr "../../db_versions/crdb/aux.version";

/**
* Uncomment to use yugabyte
local rid_db_version = importstr "../../db_versions/yugabyte/rid.version";
local scd_db_version = importstr "../../db_versions/yugabyte/scd.version";
local aux_db_version = importstr "../../db_versions/yugabyte/aux.version";
**/

local metadata = metadataBase {
  namespace: 'VAR_NAMESPACE',
  clusterName: 'VAR_CLUSTER_CONTEXT',
  single_cluster: false,
  enableScd: false, // <-- This boolean value is VAR_ENABLE_SCD
  datastore: 'VAR_DATASTORE',
  locality: 'VAR_LOCALITY',
  cockroach+: {
    image: 'VAR_CRDB_DOCKER_IMAGE_NAME',
    hostnameSuffix: 'VAR_DB_HOSTNAME_SUFFIX',
    nodeIPs: ['VAR_CRDB_NODE_IP1', 'VAR_CRDB_NODE_IP2', 'VAR_CRDB_NODE_IP3'],
    shouldInit: false, // <-- This boolean value is VAR_SHOULD_INIT
    JoinExisting: ['VAR_CRDB_EXTERNAL_NODE1', 'VAR_CRDB_EXTERNAL_NODE1', 'VAR_CRDB_EXTERNAL_NODE1' ],
    storageClass: 'VAR_STORAGE_CLASS',
  },
  yugabyte+: {
    image: 'VAR_YUGABYTE_DOCKER_IMAGE_NAME',
    storageClass: 'VAR_STORAGE_CLASS',
    masterNodeIPs: ['VAR_YUGABYTE_MASTER_IP1', 'VAR_YUGABYTE_MASTER_IP2', 'VAR_YUGABYTE_MASTER_IP3'],
    tserverNodeIPs: ['VAR_YUGABYTE_TSERVER_IP1', 'VAR_YUGABYTE_TSERVER_IP2', 'VAR_YUGABYTE_TSERVER_IP3'],
    masterAddresses: ['VAR_YUGABYTE_MASTER_ADDRESS1', 'VAR_YUGABYTE_MASTER_ADDRESS2', 'VAR_YUGABYTE_MASTER_ADDRESS3'],
    master: {
      rpc_bind_addresses: "VAR_YUGABYTE_MASTER_RPC_BIND_ADDRESSES",
      server_broadcast_addresses: "VAR_YUGABYTE_MASTER_SERVER_BROADCAST_ADDRESSES",
    },
    tserver: {
      rpc_bind_addresses: "VAR_YUGABYTE_TSERVER_RPC_BIND_ADDRESSES",
      server_broadcast_addresses: "VAR_YUGABYTE_TSERVER_SERVER_BROADCAST_ADDRESSES",
    },
    fix_27367_issue: false, // <- This boolean value is VAR_YUGABYTE_FIX_27367_ISSUE
    light_resources: false, // <- This boolean value is VAR_YUGABYTE_LIGHT_RESOURCES
    placement: {
      cloud: 'VAR_YUGABYTE_PLACEMENT_CLOUD',
      region: 'VAR_YUGABYTE_PLACEMENT_REGION',
      zone: 'VAR_YUGABYTE_PLACEMENT_ZONE',
    },
  },
  backend+: {
    ipName: 'VAR_INGRESS_NAME',
    image: 'VAR_DOCKER_IMAGE_NAME',
    pubKeys: ['VAR_PUBLIC_KEY_PEM_PATH'],
    jwksEndpoint: 'VAR_JWKS_ENDPOINT',
    jwksKeyIds: ['VAR_JWKS_KEY_ID'],
    hostname: 'VAR_APP_HOSTNAME',
    publicEndpoint: 'VAR_PUBLIC_ENDPOINT',
    dumpRequests: false,
    sslPolicy: 'VAR_SSL_POLICY'
  },
  schema_manager+: {
    enable: false, // <-- this boolean value is VAR_ENABLE_SCHEMA_MANAGER
    image: 'VAR_DOCKER_IMAGE_NAME',
    desired_rid_db_version: rid_db_version,
    desired_scd_db_version: scd_db_version,
    desired_aux_db_version: aux_db_version,
  },
  prometheus+: {
    storageClass: 'VAR_STORAGE_CLASS',
  },
//  image_pull_secret: 'VAR_DOCKER_IMAGE_PULL_SECRET'
};

dss.all(metadata)
