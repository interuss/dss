// This file shows the mimimum information required to get a DSS instance running in Kubernetes.
local prod = import 'prod.libsonnet';
local gateway = import '../../../deploy/http-gateway.libsonnet';

local metadata = prod.metadata {
  namespace: 'dss-main',
  clusterName: 'your_cluster_context',
  cockroach+: {
    hostnameSuffix: 'db.your_hostname_suffix.com',
    locality: 'your_unique_locality',
    nodeIPs: ['0.0.0.0', '1.1.1.1', '2.2.2.2'],
  },
  gateway+: {
    ipName: 'your-ingress-name',
    hostname: 'your_hostname.com',
  },
};

prod.all(metadata) {
  gateway+: {
    ingress: gateway.PresharedCertIngress(metadata, 'my-cert-name'),
  },
}
