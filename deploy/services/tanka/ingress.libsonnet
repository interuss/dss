local base = import 'base.libsonnet';

local awsLoadBalancer(metadata) = base.AWSLoadBalancerWithManagedCert(metadata, 'gateway', [metadata.backend.ipName], metadata.subnet, metadata.backend.certName) {
  app:: 'core-service',
  spec+: {
    ports: [{
      port: 443,
      targetPort: metadata.backend.port,
      protocol: "TCP",
      name: "http",
    }]
  }
};

local minikubeLoadBalancer(metadata) = base.Service(metadata, 'gateway') {
  app:: 'core-service',
  port:: metadata.backend.port,
  type:: 'LoadBalancer',
  spec+: {
    ports: [{
      port: 80,
      targetPort: metadata.backend.port,
      protocol: "TCP",
      name: "http",
    }]
  }
};

{
  GoogleIngress(metadata): base.Ingress(metadata, 'https-ingress') {
    metadata+: {
      annotations: {
        'kubernetes.io/ingress.global-static-ip-name': metadata.backend.ipName,
        'kubernetes.io/ingress.allow-http': 'false',
        [if metadata.backend.sslPolicy != '' then 'networking.gke.io/v1beta1.FrontendConfig']: 'ssl-frontend-config',
      },
    },
    spec: {
      defaultBackend: {
        service: {
          name: 'core-service',
          port: {
            number: metadata.backend.port,
          }
        }
      },
    },
  },

  GoogleSSLPolicyFrontendConfig(metadata): base.GoogleFrontendConfig(metadata, 'ssl-frontend-config', metadata.backend.sslPolicy),

  GoogleManagedCertIngress(metadata): {
    ingress: $.GoogleIngress(metadata) {
      metadata+: {
        annotations+: {
          'networking.gke.io/managed-certificates': 'https-certificate',
        },
      },
    },
    managedCert: base.ManagedCert(metadata, 'https-certificate') {
      spec: {
        domains: [
          metadata.backend.hostname,
        ],
      },
    },
  },

  GooglePresharedCertIngress(metadata, certName): $.GoogleIngress(metadata) {
    metadata+: {
      annotations+: {
        'ingress.gcp.kubernetes.io/pre-shared-cert': certName,
      },
    },
  },

  GoogleService(metadata): base.Service(metadata, 'core-service') {
    app:: 'core-service',
    port:: metadata.backend.port,
    type:: 'NodePort',
    enable_monitoring:: false,
  },

  CloudNetwork(metadata): {
    google: if metadata.cloud_provider == "google" then {
      ingress: if metadata.backend.preSharedCert != '' then $.GooglePresharedCertIngress(metadata, metadata.backend.preSharedCert) else $.GoogleManagedCertIngress(metadata),
      [if metadata.backend.sslPolicy != '' then 'frontendConfig']: $.GoogleSSLPolicyFrontendConfig(metadata),
      service: $.GoogleService(metadata),
    },
    aws_loadbalancer: if metadata.cloud_provider == "aws" then awsLoadBalancer(metadata),
    minikube_loadbalancer: if metadata.cloud_provider == "minikube" then minikubeLoadBalancer(metadata),
  },

  all(metadata): $.CloudNetwork(metadata),
}
