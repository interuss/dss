 { 
  all(metadata): {
    solver_servie: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: 'acme-http01-solver',
        namespace: 'istio-system',
      },
      spec: {
        ports: [
          {
            port: 8089,
            targetPort: 8089,
            protocol: 'TCP',
            name: 'http',
          },
        ],
        selector: {
          'acme.cert-manager.io/http01-solver': "true",
        },
      },
    },
    disable_mtls: {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: 'DestinationRule',
      metadata: {
        name: 'acme-http01-solver',
        namespace: 'istio-system',
      },
      spec: {
        host: '*.istio-system.svc.cluster.local',
        trafficPolicy: {
          tls: {
            mode: 'ISTIO_MUTUAL',
          },
          portLevelSettings: [
            {
              port: {
                number: 8089,
              },
              tls: {
                mode: 'DISABLE',
              },
            },
          ],
        },
      },
    },
    virtual_service: {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: 'VirtualService',
      metadata: {
        name: 'acme-http01-solver',
        namespace: 'istio-system',
      },
      spec: {
        gateways: [
          'istio-system/ingressgateway',
        ],
        hosts: [
          'spiffe-big-1.interussplatform.dev',
        ],
        http: [
          {
            match: [
              {
                uri: {
                  prefix: '/.well-known/acme-challenge/',
                },
              },
            ],
            route: [
              {
                destination: {
                  host: 'acme-http01-solver.istio-system.svc.cluster.local',
                  port: {
                    number: 8089,
                  },
                },
                weight: 100,
              },
            ],
          },
        ],
      },
    },
    https_issuer: {
      apiVersion: 'cert-manager.io/v1alpha2',
      kind: 'ClusterIssuer',
      metadata: {
        name: 'letsencrypt-prod',
      },
      spec: {
        acme: {
          server: 'https://acme-v02.api.letsencrypt.org/directory',
          email: 'steeling@google.com',
          privateKeySecretRef: {
            name: 'letsencrypt-prod',
          },
          solvers: [
            {
              http01: {
                ingress: {
                  class: 'istio',
                },
              },
            },
          ],
        },
      },
    },
    ca_ingress_cert: {
      apiVersion: 'cert-manager.io/v1alpha2',
      kind: 'Certificate',
      metadata: {
        name: 'istio-ingressgateway-certs',
        namespace: 'istio-system',
      },
      spec: {
        secretName: 'istio-ingressgateway-certs',
        issuerRef: {
          name: 'letsencrypt-prod',
          kind: 'ClusterIssuer',
        },
        dnsNames: [
          metadata.gateway.hostname,
        ],
        acme: {
          config: [
            {
              http01: {
                ingressClass: 'istio',
              },
              domains: [
                metadata.gateway.hostname,
              ],
            },
          ],
        },
      },
    },
    ca_issuer: {
      apiVersion: 'cert-manager.io/v1alpha2',
      kind: 'Issuer',
      metadata: {
        name: 'cockroach-external-tls-ca-issuer',
        namespace: 'cert-manager',
      },
      spec: {
        selfSigned: {},
      },
    },
    crdb_external_certificate: {
      apiVersion: 'cert-manager.io/v1alpha2',
      kind: 'Certificate',
      metadata: {
        name: 'cockroach-external-tls-cert',
        namespace: 'cert-manager',
      },
      spec: {
        secretName: 'cockroach-external-tls-cert',
        duration: '2160h',
        renewBefore: '360h',
        commonName: '*.' + metadata.cockroach.hostnameSuffix,
        isCA: false,
        keySize: 2048,
        keyAlgorithm: 'rsa',
        keyEncoding: 'pkcs1',
        usages: [
          'server auth',
          'client auth',
        ],
        dnsNames: [
          '*.' + metadata.cockroach.hostnameSuffix,
        ],
        # ipAddresses: [
        #   '192.168.0.5',
        # ],
        issuerRef: {
          name: 'cockroach-external-tls-ca-issuer',
          kind: 'Issuer',
          group: 'cert-manager.io',
        },
      },
    },
  },
}