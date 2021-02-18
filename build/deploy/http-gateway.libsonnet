local base = import 'base.libsonnet';

{

  all(metadata): {
    service: base.Service(metadata, 'http-gateway') {
      app:: 'http-gateway',
      port:: metadata.gateway.port,
      type:: 'NodePort',
      enable_monitoring:: false,
    },

    deployment: base.Deployment(metadata, 'http-gateway') {
      metadata+: {
        namespace: metadata.namespace,
      },
      spec+: {
        template+: {
          metadata+: {
            annotations+: {
              "sidecar.istio.io/inject": "true",
            },
          },
          spec+: {
            soloContainer:: base.Container('http-gateway') {
              image: metadata.gateway.image,
              ports: [
                {
                  containerPort: metadata.gateway.port,
                  name: 'http',
                },
              ],
              command: ['http-gateway'],
              args_:: {
                'grpc-backend': 'grpc-backend.' + metadata.namespace + ':' + metadata.backend.port,
                addr: ':' + metadata.gateway.port,
                'gcp_prof_service_name': metadata.gateway.prof_http_name,
                enable_scd: metadata.enableScd,
                'trace-requests': metadata.gateway.traceRequests,
              },
              readinessProbe: {
                httpGet: {
                  path: '/healthy',
                  port: metadata.gateway.port,
                },
              },
            },
          },
        },
      },
    },
  },
}
