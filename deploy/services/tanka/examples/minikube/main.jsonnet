# The path used in imports below must be updated to point to /deploy/services/tanka/

local dss = import '../dss.libsonnet';
local metadataBase = import '../metadata_base.libsonnet';

local metadata = metadataBase {
  cloud_provider: 'minikube',
  namespace: 'default',
  clusterName: 'dss-local-cluster',
  single_cluster: true,
  enableScd: true,
  datastore: 'yugabyte',
  locality: 'minikube',
  cockroach+: {
    image: 'cockroachdb/cockroach:v24.1.3',
    nodeIPs: ['', '', ''],
    shouldInit: true,
  },
  yugabyte+: {
    image: 'yugabytedb/yugabyte:2.25.1.0-b381',
    masterNodeIPs: ['', '', ''],
    tserverNodeIPs: ['', '', ''],
    placement: {
      cloud: 'cloud-1',
      region: 'uss-1',
      zone: 'zone-1',
    },
  },
  backend+: {
    ipName: 'VAR_INGRESS_NAME',
    image: 'docker.io/interuss-local/dss:latest',
    pubKeys: ['/test-certs/auth2.pem'],
    hostname: 'local',
    publicEndpoint: 'http://127.0.0.1:8888',
    dumpRequests: false,
  },
  schema_manager+: {
    enable: true,
    image: 'docker.io/interuss-local/dss:latest',
    desired_rid_db_version: '1.0.1',
    desired_scd_db_version: '1.0.1',
    desired_aux_db_version: '1.0.0',
  },
};

dss.all(metadata)
