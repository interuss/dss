local util = import 'util.libsonnet';

{
  volumes: {
    client_certs: {
      name: 'client-certs',
      secret: {
        secretName: 'cockroachdb.client.root',
        defaultMode: 256,
      },
    },
    node_certs: {
      name: 'node-certs',
      secret: {
        secretName: 'cockroachdb.node',
        defaultMode: 256,
      },
    },
    ca_certs: {
      name: 'ca-certs',
      secret: {
        secretName: 'cockroachdb.ca.crt',
        defaultMode: 256,
      },
    },
    datadir: {
      name: 'datadir',
      persistentVolumeClaim: {
        claimName: 'datadir',
      },
    },
    public_certs: {
      name: 'public-certs',
      secret: {
        secretName: 'dss.public.certs',
        defaultMode: 256,
      },
    },
  },  
  cockroachVolumes: util.mapToList(util.exclude(self.volumes, std.set(['public_certs']))),
  backendVolumes: util.mapToList(util.exclude(self.volumes, std.set(['node_certs', 'datadir']))),

  mounts: {
    datadir: [
      {
        name: 'datadir',
        mountPath: '/cockroach/cockroach-data',
      },
    ],
    nodeCert: [
      {
        name: 'node-certs',
        mountPath: '/cockroach/cockroach-certs/node.crt',
        subPath: 'node.crt',
      },
      {
        name: 'node-certs',
        mountPath: '/cockroach/cockroach-certs/node.key',
        subPath: 'node.key',
      },
    ],
    clientCert: [
      {
        name: 'client-certs',
        mountPath: '/cockroach/cockroach-certs/client.root.crt',
        subPath: 'client.root.crt',
      },
      {
        name: 'client-certs',
        mountPath: '/cockroach/cockroach-certs/client.root.key',
        subPath: 'client.root.key',
      },
    ],
    caCert: [
      {
        name: 'ca-certs',
        mountPath: '/cockroach/cockroach-certs/ca.crt',
        subPath: 'ca.crt',
      },
    ],
    publicCert: [
      {
        name: 'public-certs',
        mountPath: '/public-certs',
      },
    ],
  },
  cockroachMounts: std.flattenArrays(util.mapToList(util.exclude(self.mounts, std.set(['publicCert'])))),
  backendMounts: std.flattenArrays(util.mapToList(util.exclude(self.mounts, std.set(['datadir', 'nodeCert'])))),
}

