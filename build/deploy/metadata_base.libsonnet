{
  namespace: error 'must supply namespace',
  clusterName: error 'must supply cluster name',
  enable_istio: false,
  release: 'config',
  environment: 'dev',
  // Set this field if you don't intend to ever join this instance with others.
  // This disables inter cluster crdb<->crdb access when set to true.
  single_cluster: false,
  enableScd: false,
  cockroach: {
    locality: error 'must supply crdb locality',
    hostnameSuffix: error 'must supply a hostnameSuffix, or override in statefulset',
    shouldInit: false,  // Set this to true if you are starting a new cluster.
    grpc_port: 26257,
    http_port: 8080,
    image: 'cockroachdb/cockroach:v21.2.3',
    nodeIPs: error 'must supply the per-node ip addresses as an array',
    JoinExisting: [],
  },
  PSP: {
    roleRef: '',
    roleBinding: false,
  },
  gateway: {
    port: 8080,
    ipName: error 'must supply ip name',
    image: error 'must specify image',
    prof_http_name: '',
    hostname: error 'must specify hostname',
    traceRequests: false,
  },
  backend: {
    port: 8081,
    image: error 'must specify image',
    prof_grpc_name: '',
    pubKeys: [''],
    jwksEndpoint: '',
    jwksKeyIds: [],
  },
  alert: {
    enable: false,
    smtp: {
      host: error 'must specify smtp hostname',
      email: error 'must specify source email',
      password: error 'must specify source email password',
      dest: error 'must specify destination email',
    },
  },
  prometheus: {
    expose_external: false,
    IP: '',  // This is the static external ip address for promethus ingress, leaving blank means your cloud provider will assign an ephemeral IP
    whitelist_ip_ranges: error 'must specify whitelisted CIDR IP Blocks, or empty list for fully public access',
    retention: '15d',
    storage_size: '100Gi',
  },
}
