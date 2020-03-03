{
  all(metadata): {
    crdb_service_rule: {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: 'DestinationRule',
      metadata: {
        name: 'crdb-balanced-rule',
      },
      spec: {
        host: 'cockroachdb-balanced.' + metadata.namespace,
        trafficPolicy: {
          tls: {
            mode: 'DISABLE',
          },
        },
      },
    },
    crdb_node_rule: {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: 'DestinationRule',
      metadata: {
        name: 'crdb-svc-rule',
      },
      spec: {
        host: '*.cockroachdb.' + metadata.namespace + '.svc.cluster.local',
        trafficPolicy: {
          tls: {
            mode: 'DISABLE',
          },
        },
      },
    },
  },
}