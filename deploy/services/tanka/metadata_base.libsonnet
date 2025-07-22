{
  cloud_provider: 'google', // Either google, aws or minikube
  namespace: error 'must supply namespace',
  clusterName: error 'must supply cluster name',
  release: 'config',
  environment: 'dev',
  // Set this field if you don't intend to ever join this instance with others.
  // This disables inter cluster crdb<->crdb access when set to true.
  single_cluster: false,
  enableScd: false,
  datastore: 'cockroachdb',
  locality: error 'must supply locality',
  cockroach: {
    locality: '',
    hostnameSuffix: error 'must supply a hostnameSuffix, or override in statefulset',
    shouldInit: false,  // Set this to true if you are starting a new cluster.
    grpc_port: 26257,
    http_port: 8080,
    image: error 'must specify cockroach db image. Until DSS v0.16, the recommended CockroachDB image is `cockroachdb/cockroach:v21.2.7`. From DSS v0.17, the recommended CockroachDB image is cockroachdb/cockroach:v24.1.3. Example: cockroachdb/cockroach:v21.2.7',
    nodeIPs: error 'must supply the per-node ip addresses as an array', // For AWS, this array should contain the allocation id of the elastic ips.
    JoinExisting: [],
    storageClass: 'standard',
  },
  yugabyte: {
    image: error 'must specify yugabyte db image',
    storageClass: 'standard',
    masterNodeIPs: [],
    tserverNodeIPs: [],
    masterAddresses: ["yb-master-0.yb-masters.default.svc.cluster.local:7100", "yb-master-1.yb-masters.default.svc.cluster.local:7100", "yb-master-2.yb-masters.default.svc.cluster.local:7100"],
    placement: {
      cloud: error 'must specify placement cloud',
      region: error 'must specify placement region',
      zone: error 'must specify placement zone',
    },
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
    certName: if $.cloud_provider == "aws" then error 'must specify certName for AWS cloud provider', # Only used by AWS
    sslPolicy: '', # SSL Policy Name. Only used by Google Cloud.
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
  schema_manager: {
    enable: false, // NB: Automatically enabled if should_init is set to true.
    image: error 'must specify image',
    desired_rid_db_version: '4.0.0',
    desired_scd_db_version: '3.2.0',
  },
  image_pull_secret: '',
  subnet: if $.cloud_provider == "aws" then error 'must specify subnet for AWS cloud provider', // For AWS, subnet of the elastic ips
}
