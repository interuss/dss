{
  all(metadata): {
    crdb_service_rule: {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: 'DestinationRule',
      metadata: {
        name: 'crdb-balanced-rule',
        namespace: metadata.namespace,
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
    # GCP sends health-checks to "/" on the same port as the backend defined in the ingress
    # But whereas IStio should receive traffic on port 80, its health-check is on port 15020
    # We therefore use Istio itself to redirect the health-check
    #  - from endpoint on port 80
    #  - to endpoint on port 15020
    # health_check: {
    #   apiVersion: 'networking.istio.io/v1alpha3',
    #   kind: 'VirtualService',
    #   metadata: {
    #     name: 'health',
    #     namespace: 'istio-system',
    #   },
    #   spec: {
    #     gateways: [
    #       'http-gateway',
    #     ],
    #     hosts: [
    #       '*',
    #     ],
    #     http: [
    #       {
    #         match: [
    #           {
    #             headers: {
    #               'user-agent': {
    #                 prefix: 'GoogleHC',
    #               },
    #             },
    #             method: {
    #               exact: 'GET',
    #             },
    #             uri: {
    #               exact: '/',
    #             },
    #           },
    #         ],
    #         rewrite: {
    #           authority: 'istio-ingressgateway.istio-system.svc.cluster.local:15020',
    #           uri: '/healthz/ready',
    #         },
    #         route: [
    #           {
    #             destination: {
    #               host: 'istio-ingressgateway.istio-system.svc.cluster.local',
    #               port: {
    #                 number: 15020,
    #               },
    #             },
    #           },
    #         ],
    #       },
    #     ],
    #   },
    # },
    # dest_rule: {
    #   apiVersion: 'networking.istio.io/v1alpha3',
    #   kind: 'DestinationRule',
    #   metadata: {
    #     name: 'istio-ingressgateway',
    #     namespace: 'istio-system',
    #   },
    #   spec: {
    #     host: 'istio-ingressgateway.istio-system.svc.cluster.local',
    #     trafficPolicy: {
    #       tls: {
    #         mode: 'DISABLE',
    #       },
    #     },
    #   },
    # },
    gateway_rule: {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: 'DestinationRule',
      metadata: {
        name: 'istio-ingressgateway',
        namespace: metadata.namespace,
      },
      spec: {
        host: 'http-gateway',
        trafficPolicy: {
          tls: {
            mode: 'DISABLE',
          },
        },
      },
    },
  },
}