{
  namespace: error 'must supply namespace',
  cockroach: {
    locality: error 'must supply crdb locality',
    hostnameSuffix: error 'must supply a hostnameSuffix, or override in statefulset',
    shouldInit: false,  // Set this to true if you are starting a new cluster.
    grpc_port: 26257,
    http_port: 8080,
    image: 'cockroachdb/cockroach:v19.1.5',
    balancedIP: error 'must supply the balanced ip address',
    nodeIPs: error 'must supply the per-node ip addresses as an array',
<<<<<<< HEAD
    JoinExisting: [],
=======
>>>>>>> All files in
  },
  PSP: {
    roleRef: '',
    roleBinding: false,
  },
  gateway: {
    port: 8080,
    ipName: error 'must supply ip name',
    image: error 'must specify image',
    hostname: error 'must specify hostname',
  },
  backend: {
    port: 8081,
    image: error 'must specify image',
    pubKey: 'us-demo.pem',
  },
}
