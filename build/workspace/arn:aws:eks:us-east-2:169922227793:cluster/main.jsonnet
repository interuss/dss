local dss = import '../../../deploy/dss.libsonnet';
local metadataBase = import '../../../deploy/metadata_base.libsonnet';

// All VAR_* values below must be replaced with appropriate values; see
// dss/build/README.md for more information.

local metadata = metadataBase {
  namespace: 'default',
  clusterName: 'arn:aws:eks:us-east-2:169922227793:cluster/dss-ohio',
  enable_istio: true,
  single_cluster: false,
  // enableScd - enables strategic conflict detection functionality (currently R&D) - UTM Workflow Endpoints Operation/Constraints
  enableScd: true, // <-- This boolean value is VAR_ENABLE_SCD
  cockroach+: {
    hostnameSuffix: 'cockroach.dss-ohio.oneskysystems.com',
    locality: 'onesky_dss-ohio',
    nodeIPs: ['193.168.126.127', '193.168.45.104', '193.168.75.8'],
    shouldInit: true, // <-- This boolean value is VAR_SHOULD_INIT, set this to false unless creating the first DSS instance
    // Comment out unless joining an existing CRDB cluster (as part of an existing DSS cluster)
    // JoinExisting: ['VAR_CRDB_EXTERNAL_NODE1', 'VAR_CRDB_EXTERNAL_NODE1', 'VAR_CRDB_EXTERNAL_NODE1' ],
  },
  gateway+: {
    ipName: 'onesky-dss-ingress',
    image: '169922227793.dkr.ecr.us-east-2.amazonaws.com/onesky/dss:2021-02-03-4c5e457',
    hostname: 'dss-ohio.oneskysystems.com',
    traceRequests: true,
  },
  backend+: {
    image: '169922227793.dkr.ecr.us-east-2.amazonaws.com/onesky/dss:2021-02-03-4c5e457',
    // Provide a blank string for pubKeys if we're using JWKS (which we ought to be)
    pubKeys: [''],
    jwksEndpoint: 'https://devlogin.onesky.xyz/auth/realms/dss/protocol/openid-connect/certs',
    jwksKeyIds: ['GdXw6PWjFdEtQqrFn4_K_1wdVveH7yhADldApxKQGjM'],
  },
  schema_manager+: {
    image: '169922227793.dkr.ecr.us-east-2.amazonaws.com/onesky/db-manager:2021-02-03-4c5e457',
    desired_rid_db_version: '3.1.1',
    desired_scd_db_version: '1.0.0',
  },
};

dss.all(metadata)
