// This file shows the mimimum information required to get a DSS instance running in Kubernetes.
local dss = import '../../deploy/dss.libsonnet';
local metadataBase = import '../../deploy/metadata_base.libsonnet';

local metadata = metadataBase {
  namespace: 'your_namespace', //Same <NAMESPACE> used in the make-certs.py and apply-certs.sh scripts.
  clusterName: 'your_cluster_context', //Same <CLUSTER_CONTEXT> used in the make-certs.py, and apply-certs.sh scripts.
  cockroach+: {
    hostnameSuffix: 'db.your_hostname_suffix.com',
    locality: 'your_unique_locality',
    nodeIPs: ['0.0.0.0', '1.1.1.1', '2.2.2.2'],
    balancedIP: '3.3.3.3',
    shouldInit: true, //Set to false if joining a cluster
    JoinExisting: ['0.db.westus.example.com', '1.db.westus.example.com', '2.db.westus.example.com' ], //If joining a cluster, replace these with at least 3 FQDN's of the existing DSS cockroachdb cluster you are joining.
  },
  gateway+: {
    ipName: 'your-ingress-name', //Set this if using GKE
    image: 'your_image_name',
    hostname: 'yourhostname.com', //FQDN of your gateway ingress endpoint
  },
  backend+: {
    image: 'your_image_name',
  },
};

dss.all(metadata)
