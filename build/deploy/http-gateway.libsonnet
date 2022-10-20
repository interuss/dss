local base = import 'base.libsonnet';

local ingress(metadata) = base.Ingress(metadata, 'https-ingress') {
  metadata+: {
    annotations: {
      'kubernetes.io/ingress.global-static-ip-name': metadata.gateway.gkeIngress.ipName,
      'kubernetes.io/ingress.allow-http': 'false',
    },
  },
  spec: {
    defaultBackend: {
      service: {
        name: 'http-gateway',
        port: {
          number: metadata.gateway.port,
        }
      }
    },
  },
};

{
  GkeManagedCertIngress(metadata): {
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
    ingress: if metadata.gateway.ingress == 'gke' then
                $.GkeManagedCertIngress(metadata)
             else if metadata.gateway.ingress == 'none' then
                null
             else
                error "'metadata.gateway.ingress' should be one of 'gke', 'none'",
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
