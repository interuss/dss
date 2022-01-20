local base = import 'base.libsonnet';

local ingress(metadata) = base.Ingress(metadata, 'https-ingress') {
  metadata+: {
    annotations: {
      'kubernetes.io/ingress.global-static-ip-name': metadata.gateway.ipName,
      'kubernetes.io/ingress.allow-http': 'false',
    },
  },
  spec: {
    backend: {
      serviceName: 'http-gateway',
      servicePort: metadata.gateway.port,
    },
  },
};

{
  ManagedCertIngress(metadata): {
    ingress: ingress(metadata) {
      metadata+: {
        annotations+: {
          'networking.gke.io/managed-certificates': 'https-certificate',
        },
      },
    },
    managedCert: base.ManagedCert(metadata, 'https-certificate') {
      spec: {
        domains: [
          metadata.gateway.hostname,
        ],
      },
    },
  },

  PresharedCertIngress(metadata, certName): ingress(metadata) {
    metadata+: {
      annotations+: {
        'ingress.gcp.kubernetes.io/pre-shared-cert': certName,
      },
    },
  },


  all(metadata): {
    ingress: $.ManagedCertIngress(metadata),
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
                'core-service': 'core-service.' + metadata.namespace + ':' + metadata.backend.port,
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
