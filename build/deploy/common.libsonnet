local kube = import "kube.libsonnet"; 

{
  cockroach: {
    grpc_port: 26257,
    http_port: 8080,
    volumes_: {
      client_certs: {
        secret: {
          secretName: "cockroachdb.client.root",
          defaultMode: 256,
        }
      },
      node_certs: {
        secret: {
          secretName: "cockroachdb.node",
          defaultMode: 256,
        }
      },
      ca_certs: {
        secret: {
          secretName: "cockroachdb.ca.crt",
          defaultMode: 256,
        }
      },
      datadir: {
        persistentVolumeClaim: {
          claimName: 'datadir',
        },
      },
    },
    volumes: kube.mapToNamedList(self.volumes_),
    datadirMount: {
      name: "datadir",
      mountPath: "/cockroach/cockroach-data",
    },
    nodeCertMounts: [
      {
        name: "node-certs",
        mountPath: "/cockroach/cockroach-certs/node.crt",
        subPath: "node.crt",
      },
      {
        name: "node-certs",
        mountPath: "/cockroach/cockroach-certs/node.key",
        subPath: "node.key",
      },
    ],
    clientCertMounts: [
      {
        name: "client-certs",
        mountPath: "/cockroach/cockroach-certs/client.root.crt",
        subPath: "client.root.crt",
        },
      {
        name: "client-certs",
        mountPath: "/cockroach/cockroach-certs/client.root.key",
        subPath: "client.root.key",
      },
    ],
    caCertMount: {
      name: "ca-certs",
      mountPath: "/cockroach/cockroach-certs/ca.crt",
      subPath: "ca.crt",
    }, 

    volumeMounts: [self.datadirMount, self.caCertMount] + self.nodeCertMounts + self.clientCertMounts,
  }
}
