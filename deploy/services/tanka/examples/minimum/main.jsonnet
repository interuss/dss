
# The path used in imports below must be updated to point to /deploy/services/tanka/
local dss = import '../dss.libsonnet';
local metadataBase = import '../metadata_base.libsonnet';

// All VAR_* values below must be replaced with appropriate values; see
// dss/build/README.md for more information.

local metadata = metadataBase {
  namespace: 'VAR_NAMESPACE',
  clusterName: 'VAR_CLUSTER_CONTEXT',
  single_cluster: false,
  enableScd: false, // <-- This boolean value is VAR_ENABLE_SCD
  cockroach+: {
    image: 'VAR_CRDB_DOCKER_IMAGE_NAME',
    hostnameSuffix: 'VAR_DB_HOSTNAME_SUFFIX',
    locality: 'VAR_LOCALITY',
    nodeIPs: ['VAR_CRDB_NODE_IP1', 'VAR_CRDB_NODE_IP2', 'VAR_CRDB_NODE_IP3'],
    shouldInit: false, // <-- This boolean value is VAR_SHOULD_INIT
    JoinExisting: ['VAR_CRDB_EXTERNAL_NODE1', 'VAR_CRDB_EXTERNAL_NODE1', 'VAR_CRDB_EXTERNAL_NODE1' ],
    storageClass: 'VAR_STORAGE_CLASS',
  },
  backend+: {
    ipName: 'VAR_INGRESS_NAME',
    image: 'VAR_DOCKER_IMAGE_NAME',
    pubKeys: ['VAR_PUBLIC_KEY_PEM_PATH'],
    jwksEndpoint: 'VAR_JWKS_ENDPOINT',
    jwksKeyIds: ['VAR_JWKS_KEY_ID'],
    hostname: 'VAR_APP_HOSTNAME',
    dumpRequests: false,
    sslPolicy: 'VAR_SSL_POLICY'
  },
  schema_manager+: {
    image: 'VAR_DOCKER_IMAGE_NAME',
    desired_rid_db_version: '4.0.0',
    desired_scd_db_version: '3.2.0',
    desired_aux_db_version: '1.0.0',
  },
  prometheus+: {
    storageClass: 'VAR_STORAGE_CLASS',
  },
//  image_pull_secret: 'VAR_DOCKER_IMAGE_PULL_SECRET'
};

dss.all(metadata)
