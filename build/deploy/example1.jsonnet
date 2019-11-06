local dss = import 'dss.libsonnet';
local metadataBase = import 'metadata_base.libsonnet';

local metadata = metadataBase {
  namespace: 'dss-main',
  cockroach+: {
    hostnameSuffix: 'db.steeling-test.interussplatform.com',
    locality: 'steeling',
    nodeIPs: ['34.67.202.2', '104.197.160.82', '104.198.148.201'],
    balancedIP: '35.239.110.26',
  },
  gateway+: {
    ipName: 'steeling-ingress',
    image: 'gateway:v1',
  },
  backend+: {
    image: 'backend:v1',
  },
};

dss.all(metadata)
