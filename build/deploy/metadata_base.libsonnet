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
    image: 'cockroachdb/cockroach:v21.2.7',
    nodeIPs: error 'must supply the per-node ip addresses as an array', // For AWS, this array should contain the allocation id of the elastic ips.
    JoinExisting: [],
    storageClass: 'standard',
  },
  PSP: {
    roleRef: '',
    roleBinding: false,
  },
  backend: {
    ipName: error 'must supply ip name', // For AWS, use the elastic ip allocation id.
    port: 8080,
    image: error 'must specify image',
    prof_grpc_name: '',
    pubKeys: [''],
    jwksEndpoint: '',
    jwksKeyIds: [],
    hostname: error 'must specify hostname',
    dumpRequests: false,
    certName: if $.cloudProvider == "aws" then error 'must specify  certName for AWS cloud provider', # Only used by AWS
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
    image: 'prom/prometheus:v2.35.0',
    expose_external: false,
    IP: '',  // This is the static external ip address for promethus ingress, leaving blank means your cloud provider will assign an ephemeral IP
    whitelist_ip_ranges: error 'must specify whitelisted CIDR IP Blocks, or empty list for fully public access',
    retention: '15d',
    storage_size: '100Gi',
    storageClass: 'standard',
    scrape_interval: '5s',
    evaluation_interval: '5s',
    custom_rules: [],  // An array of Prometheus recording rules, each of which is an object with "record" and "expr" properties.
    custom_args: [], // An array of strings to pass as commandline arguments to Prometheus.
  },
  subnet: if $.cloudProvider == "aws" then error 'must specify subnet for AWS cloud provider', // For AWS, subnet of the elastic ips
  cloudProvider: 'google', // Either google or aws
}
