{
  certificate: {
    apiVersion: 'cert-manager.io/v1alpha2',
    kind: 'Certificate',
    metadata: {
      name: 'ingress-cert',
      namespace: 'istio-system',
    },
    spec: {
      secretName: 'istio-cert-manager-certs',
      issuerRef: {
        name: 'letsencrypt',
        kind: 'ClusterIssuer',
      },
      commonName: '---',
      dnsNames: [
        '---',
      ],
      acme: {
        config: [
          {
            http01: {
              ingressClass: 'istio',
            },
            domains: [
              '---',
            ],
          },
        ],
      },
    },
  },
  issuer: {
    apiVersion: 'cert-manager.io/v1alpha2',
    kind: 'ClusterIssuer',
    metadata: {
      name: 'letsencrypt',
    },
    spec: {
      acme: {
        email: 'steeling@google.com',
        privateKeySecretRef: {
          name: 'letsencrypt',
        },
        server: 'https://acme-v02.api.letsencrypt.org/directory',
        solvers: [
          {
            http01: {
              ingress: {
                class: 'istio',
                podTemplate: {
                  metadata: {
                    annotations: {
                      'sidecar.istio.io/inject': 'true',
                    },
                  },
                },
              },
            },
          },
        ],
      },
    },
  },
}
