# The path used in imports below must be updated to point to /deploy/services/tanka/

local dss = import '../dss.libsonnet';
local metadataBase = import '../metadata_base.libsonnet';

local metadata = metadataBase {
  cloud_provider: 'minikube',
  namespace: 'default',
  clusterName: 'dss-local-cluster',
  single_cluster: true,
  enableScd: true,
  cockroach+: {
    image: 'cockroachdb/cockroach:v24.1.3',
    locality: 'minikube',
    nodeIPs: ['', '', ''],
    shouldInit: true,
  },
  backend+: {
    ipName: 'VAR_INGRESS_NAME',
    image: 'docker.io/interuss-local/dss:latest',
    pubKeys: ['/test-certs/auth2.pem'],
    hostname: 'local',
    dumpRequests: false,
  },
  schema_manager+: {
    image: 'docker.io/interuss-local/dss:latest',
    desired_rid_db_version: '4.0.0',
    desired_scd_db_version: '3.2.0',
  },
};

dss.all(metadata)
