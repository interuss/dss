{
  all(metadata):
    if metadata.datastore == "cockroachdb" then {
      datastore_host: 'cockroachdb-balanced.' + metadata.namespace,
      datastore_port: metadata.cockroach.grpc_port,
      datastore_ssl_mode: 'verify-full',
      datastore_user: 'root',
      datastore_ssl_dir: '/cockroach/cockroach-certs',
    } else {
      datastore_host: 'yb-tservers.' + metadata.namespace,
      datastore_port: 5433,
      datastore_ssl_mode: 'verify-full',
      datastore_user: 'yugabyte',
      datastore_ssl_dir: '/opt/yugabyte-certs/',
    }
}
