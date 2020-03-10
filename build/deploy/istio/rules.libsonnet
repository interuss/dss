{
  all(metadata): {
    crdb_headless_service_entry: {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: 'ServiceEntry',
      metadata: {
        name: 'crdb-stateful-service-entry',
        namespace: metadata.namespace,
      },
      spec: {
        hosts: [
          '*.cockroachdb.' + metadata.namespace + '.svc.cluster.local',
          '*.cockroachdb',
        ],
        location: 'MESH_INTERNAL',
        ports: [
          {
            number: 26257,
            name: 'crdbheadless1',
            protocol: 'TCP',
          },
          {
            number: 8080,
            name: 'crdbheadless2',
            protocol: 'HTTP',
          },
        ],
        resolution: 'NONE',
      },
    },
  },
}