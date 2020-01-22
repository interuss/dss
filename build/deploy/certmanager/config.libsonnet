 { 
  all(metadata): {
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