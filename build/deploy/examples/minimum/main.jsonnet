local dss = import '../../../deploy/dss.libsonnet';
local metadataBase = import '../../../deploy/metadata_base.libsonnet';

// All VAR_* values below must be replaced with appropriate values; see
// dss/build/README.md for more information.

local metadata = metadataBase {
  namespace: 'VAR_NAMESPACE',
  clusterName: 'VAR_CLUSTER_CONTEXT',
  enable_istio: true,
  single_cluster: false,
  cockroach+: {
    hostnameSuffix: 'VAR_CRDB_HOSTNAME_SUFFIX',
    locality: 'VAR_CRDB_LOCALITY',
    nodeIPs: ['VAR_CRDB_NODE_IP1', 'VAR_CRDB_NODE_IP2', 'VAR_CRDB_NODE_IP3'],
    shouldInit: false, // <-- This boolean value is VAR_SHOULD_INIT
    JoinExisting: ['VAR_CRDB_EXTERNAL_NODE1', 'VAR_CRDB_EXTERNAL_NODE1', 'VAR_CRDB_EXTERNAL_NODE1' ],
  },
  gateway+: {
    ipName: 'VAR_INGRESS_NAME',
    image: 'VAR_DOCKER_IMAGE_NAME',
    hostname: 'VAR_APP_HOSTNAME',
  },
  backend+: {
    image: 'VAR_DOCKER_IMAGE_NAME',
    pubKeys: ['VAR_PUBLIC_KEY_PEM_PATH'],
    jwksEndpoint: 'VAR_JWKS_ENDPOINT',
    jwksKeyIds: ['VAR_JWKS_KEY_ID'],
  },
  schema_manager+: {
    image: 'your_schema_manager_image_name',
    desired_rid_db_version: 'v3.1.0'
  },
};

dss.all(metadata)
