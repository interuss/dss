{
  all(metadata):
    if metadata.datastore == "cockroachdb" then {
      cockroach_host: 'cockroachdb-balanced.' + metadata.namespace,
      cockroach_port: metadata.cockroach.grpc_port,
      cockroach_ssl_mode: 'verify-full',
      cockroach_user: 'root',
      cockroach_ssl_dir: '/cockroach/cockroach-certs',
    } else {
      cockroach_host: 'yb-tservers.' + metadata.namespace,
      cockroach_port: 5433,
      cockroach_ssl_mode: 'verify-full',
      cockroach_user: 'yugabyte',
      cockroach_ssl_dir: '/opt/yugabyte-certs/',
    }
}
