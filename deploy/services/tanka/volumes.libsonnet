local util = import 'util.libsonnet';

{
  all(metadata):
    if metadata.datastore == "cockroachdb" then {
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
      schemaVolumes: util.mapToList(util.exclude(self.volumes, std.set(['node_certs', 'datadir', 'public_certs']))),

      mounts: {
        datadir: [{
          name: 'datadir',
          mountPath: '/cockroach/cockroach-data',
        }],
        nodeCert: [{
          name: 'node-certs',
          mountPath: '/cockroach/cockroach-certs/node.crt',
          subPath: 'node.crt',
        }, {
          name: 'node-certs',
          mountPath: '/cockroach/cockroach-certs/node.key',
          subPath: 'node.key',
        }],
        clientCert: [{
          name: 'client-certs',
          mountPath: '/cockroach/cockroach-certs/client.root.crt',
          subPath: 'client.root.crt',
        }, {
          name: 'client-certs',
          mountPath: '/cockroach/cockroach-certs/client.root.key',
          subPath: 'client.root.key',
        }],
        caCert: [{
          name: 'ca-certs',
          mountPath: '/cockroach/cockroach-certs/ca.crt',
          subPath: 'ca.crt',
        }],
        publicCert: [{
          name: 'public-certs',
          mountPath: '/public-certs',
        }],
      },
      cockroachMounts: std.flattenArrays(util.mapToList(util.exclude(self.mounts, std.set(['publicCert'])))),
      backendMounts: std.flattenArrays(util.mapToList(util.exclude(self.mounts, std.set(['datadir', 'nodeCert'])))),
      schemaMounts: std.flattenArrays(util.mapToList(util.exclude(self.mounts, std.set(['datadir', 'nodeCert', 'publicCert'])))),
    } else {
      volumes: {
        ca_certs: {
          name: 'ca-certs',
          secret: {
              secretName: 'yugabyte-tls-client-cert',
              defaultMode: 256,
          },
        },
        client_certs: {
          name: 'client-certs',
          secret: {
              secretName: 'yugabyte-tls-client-cert',
              defaultMode: 256,
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
      backendVolumes: util.mapToList(self.volumes),
      schemaVolumes: util.mapToList(util.exclude(self.volumes, std.set(['public_certs']))),

      mounts: {
        ca_certs: [{
          name: 'ca-certs',
          mountPath: '/opt/yugabyte-certs/ca.crt',
          subPath: 'root.crt',
        }, {
          name: 'ca-certs',
          mountPath: '/opt/yugabyte-certs/ca-instance.crt',
          subPath: 'ca-instance.crt',
        }],
        client_certs: [{
          name: 'client-certs',
          mountPath: '/opt/yugabyte-certs/client.yugabyte.crt',
          subPath: 'yugabytedb.crt',
        }, {
          name: 'client-certs',
          mountPath: '/opt/yugabyte-certs/client.yugabyte.key',
          subPath: 'yugabytedb.key',
        }],
        publicCert: [{
          name: 'public-certs',
          mountPath: '/public-certs',
        }],
      },
      backendMounts: std.flattenArrays(util.mapToList(self.mounts)),
      schemaMounts: std.flattenArrays(util.mapToList(util.exclude(self.mounts, std.set(['publicCert'])))),
    }
}
