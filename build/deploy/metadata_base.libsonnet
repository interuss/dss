{
  namespace: error 'must supply namespace',
  clusterName: error 'must supply cluster name',
  enable_istio: false,
  applied_istio_definitions: false,
  release: 'config',
  // Set this field if you don't intend to ever join this instance with others.
  // This disables inter cluster crdb<->crdb access when set to true.
  single_cluster: false,
  cockroach: {
    locality: error 'must supply crdb locality',
    hostnameSuffix: error 'must supply a hostnameSuffix, or override in statefulset',
    shouldInit: false,  // Set this to true if you are starting a new cluster.
    grpc_port: 26257,
    http_port: 8080,
    image: 'cockroachdb/cockroach:v19.1.5',
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
  },
  backend: {
    port: 8081,
    image: error 'must specify image',
    prof_grpc_name: '',
    pubKey: 'us-demo.pem',
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
    external: false,
    IP: '',  // Leaving blank means your cloud provider will assign an ephemeral IP
    whitelist_ip_ranges: [],  // Empty list means firewall rules are open
  },
}
