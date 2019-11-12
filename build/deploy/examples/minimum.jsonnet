// This file shows the mimimum information required to get a DSS instance running in Kubernetes.
local dss = import '../dss.libsonnet';
local metadataBase = import '../metadata_base.libsonnet';

local metadata = metadataBase {
  namespace: 'dss-main',
  clusterName: 'your_cluster_context',
  cockroach+: {
    hostnameSuffix: 'db.your_hostname_suffix.com',
    locality: 'your_unique_locality',
    nodeIPs: ['0.0.0.0', '1.1.1.1', '2.2.2.2'],
    balancedIP: '3.3.3.3',
    shouldInit: true,
  },
  gateway+: {
    ipName: 'your-ingress-name',
    image: 'your_image_name',
    hostname: 'yourhostname.com',
  },
  backend+: {
    image: 'your_image_name',
  },
};

dss.all(metadata)
