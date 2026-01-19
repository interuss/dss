{
  scrape_configs: [
    {
      job_name: 'K8s-Endpoints',
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
        {
          source_labels: [
            '__meta_kubernetes_pod_label_statefulset_kubernetes_io_pod_name',
          ],
          action: 'replace',
          target_label: 'pod_name',
          regex: '(cockroachdb-\\d+)',
        },
        {
          source_labels: [
            '__meta_kubernetes_endpoints_label_app',
          ],
          regex: '(yb-.*)',
          replacement: 'dss',
          target_label: 'node_prefix',
        },
        {
          source_labels: [
            '__meta_kubernetes_endpoints_label_app',
          ],
          regex: '(yb-master.*)',
          replacement: 'master_export',
          target_label: 'export_type',
        },
        {
          source_labels: [
            '__meta_kubernetes_endpoints_label_app',
          ],
          regex: '(yb-master.*)',
          replacement: 'yb-master',
          target_label: 'group',
        },
        {
          source_labels: [
            '__meta_kubernetes_endpoints_label_app',
          ],
          regex: '(yb-tserver.*)',
          replacement: 'tserver_export',
          target_label: 'export_type',
        },
        {
          source_labels: [
            '__meta_kubernetes_endpoints_label_app',
          ],
          regex: '(yb-tserver.*)',
          replacement: 'yb-tserver',
          target_label: 'group',
        },
        {
          source_labels: [
            '__meta_kubernetes_endpoints_label_app',
          ],
          regex: '(yb-ysql.*)',
          replacement: 'ysql_export',
          target_label: 'export_type',
        },
        {
          source_labels: [
            '__meta_kubernetes_endpoints_label_app',
          ],
          regex: '(yb-ysql.*)',
          replacement: 'ysql',
          target_label: 'group',
        },
      ],
      metric_relabel_configs: [
        {
          source_labels: [
            '__name__',
          ],
          regex: "(.*)",
          target_label: "saved_name",
          replacement: "$1",
        },
        {
          source_labels: [
            '__name__',
          ],
          regex: "handler_latency_(yb_[^_]*)_([^_]*)_([^_]*)(.*)",
          target_label: "server_type",
          replacement: "$1",
        },
        {
          source_labels: [
            '__name__',
          ],
          regex: "handler_latency_(yb_[^_]*)_([^_]*)_([^_]*)(.*)",
          target_label: "service_type",
          replacement: "$2",
        },
        {
          source_labels: [
            '__name__',
          ],
          regex: "handler_latency_(yb_[^_]*)_([^_]*)_([^_]*)(_sum|_count)?",
          target_label: "service_method",
          replacement: "$3",
        },
        {
          source_labels: [
            '__name__',
          ],
          regex: "handler_latency_(yb_[^_]*)_([^_]*)_([^_]*)(_sum|_count)?",
          target_label: "__name__",
          replacement: "rpc_latency$4",
        },
      ],
      tls_config: {
        insecure_skip_verify: true,
      },
    },
  ],
}
