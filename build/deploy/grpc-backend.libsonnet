local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';

{
  all(metadata): {
    service: base.Service(metadata, 'grpc-backend') {
      app:: 'grpc-backend',
      port:: metadata.backend.port,
      enable_monitoring:: true,
    },

    deployment: base.Deployment(metadata, 'grpc-backend') {
      apiVersion: 'apps/v1beta1',
      kind: 'Deployment',
      app:: 'grpc-backend',
      metadata+: {
        namespace: metadata.namespace,
      },
      spec+: {
        replicas: 1,
        template+: {
          spec+: {
            volumes: volumes.backendVolumes,
            soloContainer:: base.Container('grpc-backend') {
              image: metadata.backend.image,
              imagePullPolicy: 'IfNotPresent',
              ports: [
                {
                  containerPort: metadata.backend.port,
                  name: 'grpc',
                },
              ],
              volumeMounts: volumes.backendMounts,
              command: ['grpc-backend'],
              args_:: {
                addr: ':' + metadata.backend.port,
                cockroach_host: 'cockroachdb-balanced.' + metadata.namespace,
                cockroach_port: metadata.cockroach.grpc_port,
                cockroach_ssl_mode: 'verify-full',
                cockroach_user: 'root',
<<<<<<< HEAD
<<<<<<< HEAD
                cockroach_ssl_dir: '/cockroach/cockroach-certs',
=======
                cockroach_ssl_dir: '/cockroach-certs',
>>>>>>> All files in
=======
                cockroach_ssl_dir: '/cockroach/cockroach-certs',
>>>>>>> jsonnet and kubecfg documentation
                public_key_file: '/public-certs/' + metadata.backend.pubKey,
                dump_requests: true,
                jwt_audience: metadata.gateway.hostname,
              },
            },
          },
        },
      },
    },
  },
}
