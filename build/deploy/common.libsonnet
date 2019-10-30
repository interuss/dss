local kube = import "kube.libsonnet"; 

{
  cockroach: {
    grpc_port: 26257,
    http_port: 8080,
    volumes_: {
      client_certs: {
        secretName: "cockroachdb.client.root",
        defaultMode: 256,
      },
      node_certs: {
        secretName: "cockroachdb.node",
        defaultMode: 256,
      },
    },
    volumes: kube.mapToNamedList(volumes_),
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

          # {
          #   name: 'datadir',
          #   persistentVolumeClaim: {
          #     claimName: 'datadir',
          #   },
          # },
          # {
          #   name: 'certs',
          #   secret: {
          #     secretName: 'cockroachdb.node',
          #     defaultMode: 256,
          #   },
          # },