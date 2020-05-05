local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';

{
  all(metadata): {
    service: base.Service(metadata, 'grpc-backend') {
      app:: 'grpc-backend',
      port:: metadata.backend.port,
      enable_monitoring:: false,
    },

    deployment: base.Deployment(metadata, 'grpc-backend') {
      apiVersion: 'apps/v1beta1',
      kind: 'Deployment',
      app:: 'grpc-backend',
      metadata+: {
        namespace: metadata.namespace,
      },
      spec+: {
        template+: {
          spec+: {
            volumes: volumes.backendVolumes,
            soloContainer:: base.Container('grpc-backend') {
              image: metadata.backend.image,
              imagePullPolicy: 'Always',
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
                gcp_prof_service_name: metadata.backend.prof_grpc_name,
                cockroach_host: 'cockroachdb-balanced.' + metadata.namespace,
                cockroach_port: metadata.cockroach.grpc_port,
                cockroach_ssl_mode: 'verify-full',
                cockroach_user: 'root',
                cockroach_ssl_dir: '/cockroach/cockroach-certs',
                public_key_files: std.join(metadata.backend.pubKeys, ","),
                dump_requests: true,
                accepted_jwt_audiences: metadata.gateway.hostname,
              },
            },
          },
        },
      },
    },
  },
}
