local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';

{
  all(metadata): {
    service: base.Service(metadata, 'core-service') {
      app:: 'core-service',
      port:: metadata.backend.port,
      enable_monitoring:: false,
    },

    deployment: base.Deployment(metadata, 'core-service') {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata+: {
        namespace: metadata.namespace,
      },
      spec+: {
        template+: {
          spec+: {
            volumes: volumes.backendVolumes,
            soloContainer:: base.Container('core-service') {
              image: metadata.backend.image,
              imagePullPolicy: 'Always',
              ports: [
                {
                  containerPort: metadata.backend.port,
                  name: 'grpc',
                },
              ],
              volumeMounts: volumes.backendMounts,
              command: ['core-service'],
              args_:: {
                addr: ':' + metadata.backend.port,
                gcp_prof_service_name: metadata.backend.prof_grpc_name,
                cockroach_host: 'cockroachdb-balanced.' + metadata.namespace,
                cockroach_port: metadata.cockroach.grpc_port,
                cockroach_ssl_mode: 'verify-full',
                cockroach_user: 'root',
                cockroach_ssl_dir: '/cockroach/cockroach-certs',
                public_key_files: std.join(",", metadata.backend.pubKeys),
                jwks_endpoint: metadata.backend.jwksEndpoint,
                jwks_key_ids: std.join(",", metadata.backend.jwksKeyIds),
                dump_requests: true,
                accepted_jwt_audiences: metadata.gateway.hostname,
                locality: metadata.cockroach.locality,
                enable_scd: metadata.enableScd,
              },
            },
          },
        },
      },
    },
  },
}
