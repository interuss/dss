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
  ],
  tls_config: {
    insecure_skip_verify: true,
  },
}