{
  scrape_configs: [
    {
      job_name: 'istio-mesh',
      kubernetes_sd_configs: [
        {
          role: 'endpoints',
          namespaces: {
            names: [
              'istio-system',
            ],
          },
        },
      ],
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_service_name',
            '__meta_kubernetes_endpoint_port_name',
          ],
          action: 'keep',
          regex: 'istio-telemetry;prometheus',
        },
      ],
    },
    {
      job_name: 'envoy-stats',
      metrics_path: '/stats/prometheus',
      kubernetes_sd_configs: [
        {
          role: 'pod',
        },
      ],
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_pod_container_port_name',
          ],
          action: 'keep',
          regex: '.*-envoy-prom',
        },
        {
          source_labels: [
            '__address__',
            '__meta_kubernetes_pod_annotation_prometheus_io_port',
          ],
          action: 'replace',
          regex: '([^:]+)(?::\\d+)?;(\\d+)',
          replacement: '$1:15090',
          target_label: '__address__',
        },
        {
          action: 'labelmap',
          regex: '__meta_kubernetes_pod_label_(.+)',
        },
        {
          source_labels: [
            '__meta_kubernetes_namespace',
          ],
          action: 'replace',
          target_label: 'namespace',
        },
        {
          source_labels: [
            '__meta_kubernetes_pod_name',
          ],
          action: 'replace',
          target_label: 'pod_name',
        },
      ],
    },
    {
      job_name: 'istio-policy',
      kubernetes_sd_configs: [
        {
          role: 'endpoints',
          namespaces: {
            names: [
              'istio-system',
            ],
          },
        },
      ],
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_service_name',
            '__meta_kubernetes_endpoint_port_name',
          ],
          action: 'keep',
          regex: 'istio-policy;http-policy-monitoring',
        },
      ],
    },
    {
      job_name: 'istio-telemetry',
      kubernetes_sd_configs: [
        {
          role: 'endpoints',
          namespaces: {
            names: [
              'istio-system',
            ],
          },
        },
      ],
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_service_name',
            '__meta_kubernetes_endpoint_port_name',
          ],
          action: 'keep',
          regex: 'istio-telemetry;http-monitoring',
        },
      ],
    },
    {
      job_name: 'pilot',
      kubernetes_sd_configs: [
        {
          role: 'endpoints',
          namespaces: {
            names: [
              'istio-system',
            ],
          },
        },
      ],
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_service_name',
            '__meta_kubernetes_endpoint_port_name',
          ],
          action: 'keep',
          regex: 'istio-pilot;http-monitoring',
        },
      ],
    },
    {
      job_name: 'galley',
      kubernetes_sd_configs: [
        {
          role: 'endpoints',
          namespaces: {
            names: [
              'istio-system',
            ],
          },
        },
      ],
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_service_name',
            '__meta_kubernetes_endpoint_port_name',
          ],
          action: 'keep',
          regex: 'istio-galley;http-monitoring',
        },
      ],
    },
    {
      job_name: 'citadel',
      kubernetes_sd_configs: [
        {
          role: 'endpoints',
          namespaces: {
            names: [
              'istio-system',
            ],
          },
        },
      ],
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_service_name',
            '__meta_kubernetes_endpoint_port_name',
          ],
          action: 'keep',
          regex: 'istio-citadel;http-monitoring',
        },
      ],
    },
    {
      job_name: 'kubernetes-apiservers',
      kubernetes_sd_configs: [
        {
          role: 'endpoints',
          namespaces: {
            names: [
              'default',
            ],
          },
        },
      ],
      scheme: 'https',
      tls_config: {
        ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
      },
      bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_service_name',
            '__meta_kubernetes_endpoint_port_name',
          ],
          action: 'keep',
          regex: 'kubernetes;https',
        },
      ],
    },
    {
      job_name: 'kubernetes-nodes',
      scheme: 'https',
      tls_config: {
        ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
      },
      bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',
      kubernetes_sd_configs: [
        {
          role: 'node',
        },
      ],
      relabel_configs: [
        {
          action: 'labelmap',
          regex: '__meta_kubernetes_node_label_(.+)',
        },
        {
          target_label: '__address__',
          replacement: 'kubernetes.default.svc:443',
        },
        {
          source_labels: [
            '__meta_kubernetes_node_name',
          ],
          regex: '(.+)',
          target_label: '__metrics_path__',
          replacement: '/api/v1/nodes/${1}/proxy/metrics',
        },
      ],
    },
    {
      job_name: 'kubernetes-cadvisor',
      scheme: 'https',
      tls_config: {
        ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt',
      },
      bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token',
      kubernetes_sd_configs: [
        {
          role: 'node',
        },
      ],
      relabel_configs: [
        {
          action: 'labelmap',
          regex: '__meta_kubernetes_node_label_(.+)',
        },
        {
          target_label: '__address__',
          replacement: 'kubernetes.default.svc:443',
        },
        {
          source_labels: [
            '__meta_kubernetes_node_name',
          ],
          regex: '(.+)',
          target_label: '__metrics_path__',
          replacement: '/api/v1/nodes/${1}/proxy/metrics/cadvisor',
        },
      ],
    },
    {
      job_name: 'kubernetes-service-endpoints',
      kubernetes_sd_configs: [
        {
          role: 'endpoints',
        },
      ],
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_service_annotation_prometheus_io_scrape',
          ],
          action: 'keep',
          regex: true,
        },
        {
          source_labels: [
            '__meta_kubernetes_service_annotation_prometheus_io_scheme',
          ],
          action: 'replace',
          target_label: '__scheme__',
          regex: '(https?)',
        },
        {
          source_labels: [
            '__meta_kubernetes_service_annotation_prometheus_io_path',
          ],
          action: 'replace',
          target_label: '__metrics_path__',
          regex: '(.+)',
        },
        {
          source_labels: [
            '__address__',
            '__meta_kubernetes_service_annotation_prometheus_io_port',
          ],
          action: 'replace',
          target_label: '__address__',
          regex: '([^:]+)(?::\\d+)?;(\\d+)',
          replacement: '$1:$2',
        },
        {
          action: 'labelmap',
          regex: '__meta_kubernetes_service_label_(.+)',
        },
        {
          source_labels: [
            '__meta_kubernetes_namespace',
          ],
          action: 'replace',
          target_label: 'kubernetes_namespace',
        },
        {
          source_labels: [
            '__meta_kubernetes_service_name',
          ],
          action: 'replace',
          target_label: 'kubernetes_name',
        },
      ],
    },
    {
      job_name: 'kubernetes-pods',
      kubernetes_sd_configs: [
        {
          role: 'pod',
        },
      ],
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_pod_annotation_prometheus_io_scrape',
          ],
          action: 'keep',
          regex: true,
        },
        {
          source_labels: [
            '__meta_kubernetes_pod_annotation_sidecar_istio_io_status',
          ],
          action: 'drop',
          regex: '(.+)',
        },
        {
          source_labels: [
            '__meta_kubernetes_pod_annotation_istio_mtls',
          ],
          action: 'drop',
          regex: '(true)',
        },
        {
          source_labels: [
            '__meta_kubernetes_pod_annotation_prometheus_io_path',
          ],
          action: 'replace',
          target_label: '__metrics_path__',
          regex: '(.+)',
        },
        {
          source_labels: [
            '__address__',
            '__meta_kubernetes_pod_annotation_prometheus_io_port',
          ],
          action: 'replace',
          regex: '([^:]+)(?::\\d+)?;(\\d+)',
          replacement: '$1:$2',
          target_label: '__address__',
        },
        {
          action: 'labelmap',
          regex: '__meta_kubernetes_pod_label_(.+)',
        },
        {
          source_labels: [
            '__meta_kubernetes_namespace',
          ],
          action: 'replace',
          target_label: 'namespace',
        },
        {
          source_labels: [
            '__meta_kubernetes_pod_name',
          ],
          action: 'replace',
          target_label: 'pod_name',
        },
      ],
    },
    {
      job_name: 'kubernetes-pods-istio-secure',
      scheme: 'https',
      tls_config: {
        ca_file: '/etc/istio-certs/root-cert.pem',
        cert_file: '/etc/istio-certs/cert-chain.pem',
        key_file: '/etc/istio-certs/key.pem',
        insecure_skip_verify: true,
      },
      kubernetes_sd_configs: [
        {
          role: 'pod',
        },
      ],
      relabel_configs: [
        {
          source_labels: [
            '__meta_kubernetes_pod_annotation_prometheus_io_scrape',
          ],
          action: 'keep',
          regex: true,
        },
        {
          source_labels: [
            '__meta_kubernetes_pod_annotation_sidecar_istio_io_status',
            '__meta_kubernetes_pod_annotation_istio_mtls',
          ],
          action: 'keep',
          regex: '(([^;]+);([^;]*))|(([^;]*);(true))',
        },
        {
          source_labels: [
            '__meta_kubernetes_pod_annotation_prometheus_io_path',
          ],
          action: 'replace',
          target_label: '__metrics_path__',
          regex: '(.+)',
        },
        {
          source_labels: [
            '__address__',
          ],
          action: 'keep',
          regex: '([^:]+):(\\d+)',
        },
        {
          source_labels: [
            '__address__',
            '__meta_kubernetes_pod_annotation_prometheus_io_port',
          ],
          action: 'replace',
          regex: '([^:]+)(?::\\d+)?;(\\d+)',
          replacement: '$1:$2',
          target_label: '__address__',
        },
        {
          action: 'labelmap',
          regex: '__meta_kubernetes_pod_label_(.+)',
        },
        {
          source_labels: [
            '__meta_kubernetes_namespace',
          ],
          action: 'replace',
          target_label: 'namespace',
        },
        {
          source_labels: [
            '__meta_kubernetes_pod_name',
          ],
          action: 'replace',
          target_label: 'pod_name',
        },
      ],
    },
  ],
}