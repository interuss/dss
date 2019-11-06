{
  cockroach: {
    grpc_port: 26257,
    http_port: 8080,
    volumes: {
      client_certs: {
        name: "client-certs",
        secretName: "cockroachdb.client.root",
        defaultMode: 256,
      },
      node_certs: {
        name: "node-certs",
        secretName: "cockroachdb.node",
        defaultMode: 256,
      },
    },
    volumeMounts: {
      client_certs: {
        name: "client-certs",
        secret: {
          secretName: "cockroachdb.client.root",
          defaultMode: 256,
        },
      },
    }
  }
}