local base = import 'base.libsonnet'; 

{
  all(metadata) : {
    clusterRole: base.ClusterRole(metadata, 'prometheus') {
      rules: [
        {
          apiGroups: [
            '',
          ],
          resources: [
            'nodes',
            'nodes/proxy',
            'services',
            'endpoints',
            'pods',
          ],
          verbs: [
            'get',
            'list',
            'watch',
            'create',
          ],
        },
        {
          apiGroups: [
            'extensions',
          ],
          resources: [
            'ingresses',
          ],
          verbs: [
            'get',
            'list',
            'watch',
          ],
        },
        {
          nonResourceURLs: [
            '/metrics',
            '/_status/vars',
          ],
          verbs: [
            'get',
          ],
        },
      ],
    },
    clusterRoleBinding: base.ClusterRoleBinding(metadata, 'prometheus') {
      roleRef: {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'ClusterRole',
        name: 'prometheus',
      },
      subjects: [
        {
          kind: 'ServiceAccount',
          name: 'default',
          namespace: metadata.namespace,
        },
      ],
    },
    configMap: base.ConfigMap(metadata, 'prometheus-server-conf') {
      data: {
<<<<<<< HEAD
        'prometheus.yml': "global:\n  scrape_interval: 5s\n  evaluation_interval: 5s\n\nrule_files:\n- 'aggregation.rules.yml'\n\nscrape_configs:\n\n  - job_name: 'K8s-Endpoints'\n    kubernetes_sd_configs:\n    - role: endpoints\n    relabel_configs:\n    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]\n      action: keep\n      regex: true\n    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]\n      action: replace\n      target_label: __scheme__\n      regex: (https?)\n    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]\n      action: replace\n      target_label: __metrics_path__\n      regex: (.+)\n    - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]\n      action: replace\n      target_label: __address__\n      regex: ([^:]+)(?::\\d+)?;(\\d+)\n      replacement: $1:$2\n    - action: labelmap\n      regex: __meta_kubernetes_service_label_(.+)\n    - source_labels: [__meta_kubernetes_namespace]\n      action: replace\n      target_label: kubernetes_namespace\n    - source_labels: [__meta_kubernetes_service_name]\n      action: replace\n      target_label: kubernetes_name\n    - source_labels: [__meta_kubernetes_pod_label_statefulset_kubernetes_io_pod_name]\n      action: replace\n      target_label: pod_name\n      regex: (cockroachdb-\\d+)\n    tls_config:\n      insecure_skip_verify: true",
        'aggregation.rules.yml': 'groups:\n- name: rules/aggregation.rules\n  rules:\n  - record: node:capacity\n    expr: sum without(store) (capacity{app="cockroachdb"})\n  - record: cluster:capacity\n    expr: sum without(instance) (node:capacity{app="cockroachdb"})\n  - record: node:capacity_available\n    expr: sum without(store) (capacity_available{app="cockroachdb"})\n  - record: cluster:capacity_available\n    expr: sum without(instance) (node:capacity_available{app="cockroachdb"})\n  - record: capacity_available:ratio\n    expr: capacity_available{app="cockroachdb"} / capacity{app="cockroachdb"}\n  - record: node:capacity_available:ratio\n    expr: node:capacity_available{app="cockroachdb"} / node:capacity{app="cockroachdb"}\n  - record: cluster:capacity_available:ratio\n    expr: cluster:capacity_available{app="cockroachdb"} / cluster:capacity{app="cockroachdb"}\n  # Histogram rules: these are fairly expensive to compute live, so we precompute a few percetiles.\n  - record: txn_durations_bucket:rate1m\n    expr: rate(txn_durations_bucket{app="cockroachdb"}[1m])\n  - record: txn_durations:rate1m:quantile_50\n    expr: histogram_quantile(0.5, txn_durations_bucket:rate1m)\n  - record: txn_durations:rate1m:quantile_75\n    expr: histogram_quantile(0.75, txn_durations_bucket:rate1m)\n  - record: txn_durations:rate1m:quantile_90\n    expr: histogram_quantile(0.9, txn_durations_bucket:rate1m)\n  - record: txn_durations:rate1m:quantile_95\n    expr: histogram_quantile(0.95, txn_durations_bucket:rate1m)\n  - record: txn_durations:rate1m:quantile_99\n    expr: histogram_quantile(0.99, txn_durations_bucket:rate1m)\n  - record: exec_latency_bucket:rate1m\n    expr: rate(exec_latency_bucket{app="cockroachdb"}[1m])\n  - record: exec_latency:rate1m:quantile_50\n    expr: histogram_quantile(0.5, exec_latency_bucket:rate1m)\n  - record: exec_latency:rate1m:quantile_75\n    expr: histogram_quantile(0.75, exec_latency_bucket:rate1m)\n  - record: exec_latency:rate1m:quantile_90\n    expr: histogram_quantile(0.9, exec_latency_bucket:rate1m)\n  - record: exec_latency:rate1m:quantile_95\n    expr: histogram_quantile(0.95, exec_latency_bucket:rate1m)\n  - record: exec_latency:rate1m:quantile_99\n    expr: histogram_quantile(0.99, exec_latency_bucket:rate1m)\n  - record: round_trip_latency_bucket:rate1m\n    expr: rate(round_trip_latency_bucket{app="cockroachdb"}[1m])\n  - record: round_trip_latency:rate1m:quantile_50\n    expr: histogram_quantile(0.5, round_trip_latency_bucket:rate1m)\n  - record: round_trip_latency:rate1m:quantile_75\n    expr: histogram_quantile(0.75, round_trip_latency_bucket:rate1m)\n  - record: round_trip_latency:rate1m:quantile_90\n    expr: histogram_quantile(0.9, round_trip_latency_bucket:rate1m)\n  - record: round_trip_latency:rate1m:quantile_95\n    expr: histogram_quantile(0.95, round_trip_latency_bucket:rate1m)\n  - record: round_trip_latency:rate1m:quantile_99\n    expr: histogram_quantile(0.99, round_trip_latency_bucket:rate1m)\n  - record: sql_exec_latency_bucket:rate1m\n    expr: rate(sql_exec_latency_bucket{app="cockroachdb"}[1m])\n  - record: sql_exec_latency:rate1m:quantile_50\n    expr: histogram_quantile(0.5, sql_exec_latency_bucket:rate1m)\n  - record: sql_exec_latency:rate1m:quantile_75\n    expr: histogram_quantile(0.75, sql_exec_latency_bucket:rate1m)\n  - record: sql_exec_latency:rate1m:quantile_90\n    expr: histogram_quantile(0.9, sql_exec_latency_bucket:rate1m)\n  - record: sql_exec_latency:rate1m:quantile_95\n    expr: histogram_quantile(0.95, sql_exec_latency_bucket:rate1m)\n  - record: sql_exec_latency:rate1m:quantile_99\n    expr: histogram_quantile(0.99, sql_exec_latency_bucket:rate1m)\n  - record: raft_process_logcommit_latency_bucket:rate1m\n    expr: rate(raft_process_logcommit_latency_bucket{app="cockroachdb"}[1m])\n  - record: raft_process_logcommit_latency:rate1m:quantile_50\n    expr: histogram_quantile(0.5, raft_process_logcommit_latency_bucket:rate1m)\n  - record: raft_process_logcommit_latency:rate1m:quantile_75\n    expr: histogram_quantile(0.75, raft_process_logcommit_latency_bucket:rate1m)\n  - record: raft_process_logcommit_latency:rate1m:quantile_90\n    expr: histogram_quantile(0.9, raft_process_logcommit_latency_bucket:rate1m)\n  - record: raft_process_logcommit_latency:rate1m:quantile_95\n    expr: histogram_quantile(0.95, raft_process_logcommit_latency_bucket:rate1m)\n  - record: raft_process_logcommit_latency:rate1m:quantile_99\n    expr: histogram_quantile(0.99, raft_process_logcommit_latency_bucket:rate1m)\n  - record: raft_process_commandcommit_latency_bucket:rate1m\n    expr: rate(raft_process_commandcommit_latency_bucket{app="cockroachdb"}[1m])\n  - record: raft_process_commandcommit_latency:rate1m:quantile_50\n    expr: histogram_quantile(0.5, raft_process_commandcommit_latency_bucket:rate1m)\n  - record: raft_process_commandcommit_latency:rate1m:quantile_75\n    expr: histogram_quantile(0.75, raft_process_commandcommit_latency_bucket:rate1m)\n  - record: raft_process_commandcommit_latency:rate1m:quantile_90\n    expr: histogram_quantile(0.9, raft_process_commandcommit_latency_bucket:rate1m)\n  - record: raft_process_commandcommit_latency:rate1m:quantile_95\n    expr: histogram_quantile(0.95, raft_process_commandcommit_latency_bucket:rate1m)\n  - record: raft_process_commandcommit_latency:rate1m:quantile_99\n    expr: histogram_quantile(0.99, raft_process_commandcommit_latency_bucket:rate1m)\n'
=======
        'prometheus.yml': "global:\n  scrape_interval: 15s\nscrape_configs:\n\n# Mixer scrapping. Defaults to Prometheus and mixer on same namespace.\n#\n- job_name: 'istio-mesh'\n  kubernetes_sd_configs:\n  - role: endpoints\n    namespaces:\n      names:\n      - istio-system\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]\n    action: keep\n    regex: istio-telemetry;prometheus\n\n# Scrape config for envoy stats\n- job_name: 'envoy-stats'\n  metrics_path: /stats/prometheus\n  kubernetes_sd_configs:\n  - role: pod\n\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_pod_container_port_name]\n    action: keep\n    regex: '.*-envoy-prom'\n  - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]\n    action: replace\n    regex: ([^:]+)(?::\\d+)?;(\\d+)\n    replacement: $1:15090\n    target_label: __address__\n  - action: labelmap\n    regex: __meta_kubernetes_pod_label_(.+)\n  - source_labels: [__meta_kubernetes_namespace]\n    action: replace\n    target_label: namespace\n  - source_labels: [__meta_kubernetes_pod_name]\n    action: replace\n    target_label: pod_name\n\n- job_name: 'istio-policy'\n  kubernetes_sd_configs:\n  - role: endpoints\n    namespaces:\n      names:\n      - istio-system\n\n\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]\n    action: keep\n    regex: istio-policy;http-policy-monitoring\n\n- job_name: 'istio-telemetry'\n  kubernetes_sd_configs:\n  - role: endpoints\n    namespaces:\n      names:\n      - istio-system\n\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]\n    action: keep\n    regex: istio-telemetry;http-monitoring\n\n- job_name: 'pilot'\n  kubernetes_sd_configs:\n  - role: endpoints\n    namespaces:\n      names:\n      - istio-system\n\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]\n    action: keep\n    regex: istio-pilot;http-monitoring\n\n- job_name: 'galley'\n  kubernetes_sd_configs:\n  - role: endpoints\n    namespaces:\n      names:\n      - istio-system\n\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]\n    action: keep\n    regex: istio-galley;http-monitoring\n\n- job_name: 'citadel'\n  kubernetes_sd_configs:\n  - role: endpoints\n    namespaces:\n      names:\n      - istio-system\n\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]\n    action: keep\n    regex: istio-citadel;http-monitoring\n\n# scrape config for API servers\n- job_name: 'kubernetes-apiservers'\n  kubernetes_sd_configs:\n  - role: endpoints\n    namespaces:\n      names:\n      - default\n  scheme: https\n  tls_config:\n    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt\n  bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]\n    action: keep\n    regex: kubernetes;https\n\n# scrape config for nodes (kubelet)\n- job_name: 'kubernetes-nodes'\n  scheme: https\n  tls_config:\n    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt\n  bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token\n  kubernetes_sd_configs:\n  - role: node\n  relabel_configs:\n  - action: labelmap\n    regex: __meta_kubernetes_node_label_(.+)\n  - target_label: __address__\n    replacement: kubernetes.default.svc:443\n  - source_labels: [__meta_kubernetes_node_name]\n    regex: (.+)\n    target_label: __metrics_path__\n    replacement: /api/v1/nodes/${1}/proxy/metrics\n\n# Scrape config for Kubelet cAdvisor.\n#\n# This is required for Kubernetes 1.7.3 and later, where cAdvisor metrics\n# (those whose names begin with 'container_') have been removed from the\n# Kubelet metrics endpoint.  This job scrapes the cAdvisor endpoint to\n# retrieve those metrics.\n#\n# In Kubernetes 1.7.0-1.7.2, these metrics are only exposed on the cAdvisor\n# HTTP endpoint; use \"replacement: /api/v1/nodes/${1}:4194/proxy/metrics\"\n# in that case (and ensure cAdvisor's HTTP server hasn't been disabled with\n# the --cadvisor-port=0 Kubelet flag).\n#\n# This job is not necessary and should be removed in Kubernetes 1.6 and\n# earlier versions, or it will cause the metrics to be scraped twice.\n- job_name: 'kubernetes-cadvisor'\n  scheme: https\n  tls_config:\n    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt\n  bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token\n  kubernetes_sd_configs:\n  - role: node\n  relabel_configs:\n  - action: labelmap\n    regex: __meta_kubernetes_node_label_(.+)\n  - target_label: __address__\n    replacement: kubernetes.default.svc:443\n  - source_labels: [__meta_kubernetes_node_name]\n    regex: (.+)\n    target_label: __metrics_path__\n    replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor\n\n# scrape config for service endpoints.\n- job_name: 'kubernetes-service-endpoints'\n  kubernetes_sd_configs:\n  - role: endpoints\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]\n    action: keep\n    regex: true\n  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]\n    action: replace\n    target_label: __scheme__\n    regex: (https?)\n  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]\n    action: replace\n    target_label: __metrics_path__\n    regex: (.+)\n  - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]\n    action: replace\n    target_label: __address__\n    regex: ([^:]+)(?::\\d+)?;(\\d+)\n    replacement: $1:$2\n  - action: labelmap\n    regex: __meta_kubernetes_service_label_(.+)\n  - source_labels: [__meta_kubernetes_namespace]\n    action: replace\n    target_label: kubernetes_namespace\n  - source_labels: [__meta_kubernetes_service_name]\n    action: replace\n    target_label: kubernetes_name\n\n- job_name: 'K8s-Endpoints'\n  kubernetes_sd_configs:\n  - role: endpoints\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]\n    action: keep\n    regex: true\n  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]\n    action: replace\n    target_label: __scheme__\n    regex: (https?)\n  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]\n    action: replace\n    target_label: __metrics_path__\n    regex: (.+)\n  - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]\n    action: replace\n    target_label: __address__\n    regex: ([^:]+)(?::\\d+)?;(\\d+)\n    replacement: $1:$2\n  - action: labelmap\n    regex: __meta_kubernetes_service_label_(.+)\n  - source_labels: [__meta_kubernetes_namespace]\n    action: replace\n    target_label: kubernetes_namespace\n  - source_labels: [__meta_kubernetes_service_name]\n    action: replace\n    target_label: kubernetes_name\n  - source_labels: [__meta_kubernetes_pod_label_statefulset_kubernetes_io_pod_name]\n    action: replace\n    target_label: pod_name\n    regex: (cockroachdb-\\\\d+)\n  tls_config:\n    insecure_skip_verify: true\n\n- job_name: 'kubernetes-pods'\n  kubernetes_sd_configs:\n  - role: pod\n  relabel_configs:  # If first two labels are present, pod should be scraped  by the istio-secure job.\n  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]\n    action: keep\n    regex: true\n  - source_labels: [__meta_kubernetes_pod_annotation_sidecar_istio_io_status]\n    action: drop\n    regex: (.+)\n  - source_labels: [__meta_kubernetes_pod_annotation_istio_mtls]\n    action: drop\n    regex: (true)\n  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]\n    action: replace\n    target_label: __metrics_path__\n    regex: (.+)\n  - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]\n    action: replace\n    regex: ([^:]+)(?::\\d+)?;(\\d+)\n    replacement: $1:$2\n    target_label: __address__\n  - action: labelmap\n    regex: __meta_kubernetes_pod_label_(.+)\n  - source_labels: [__meta_kubernetes_namespace]\n    action: replace\n    target_label: namespace\n  - source_labels: [__meta_kubernetes_pod_name]\n    action: replace\n    target_label: pod_name\n- job_name: 'kubernetes-pods-istio-secure'\n  scheme: https\n  tls_config:\n    ca_file: /etc/istio-certs/root-cert.pem\n    cert_file: /etc/istio-certs/cert-chain.pem\n    key_file: /etc/istio-certs/key.pem\n    insecure_skip_verify: true  # prometheus does not support secure naming.\n  kubernetes_sd_configs:\n  - role: pod\n  relabel_configs:\n  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]\n    action: keep\n    regex: true\n  # sidecar status annotation is added by sidecar injector and\n  # istio_workload_mtls_ability can be specifically placed on a pod to indicate its ability to receive mtls traffic.\n  - source_labels: [__meta_kubernetes_pod_annotation_sidecar_istio_io_status, __meta_kubernetes_pod_annotation_istio_mtls]\n    action: keep\n    regex: (([^;]+);([^;]*))|(([^;]*);(true))\n  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]\n    action: replace\n    target_label: __metrics_path__\n    regex: (.+)\n  - source_labels: [__address__]  # Only keep address that is host:port\n    action: keep    # otherwise an extra target with ':443' is added for https scheme\n    regex: ([^:]+):(\\d+)\n  - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]\n    action: replace\n    regex: ([^:]+)(?::\\d+)?;(\\d+)\n    replacement: $1:$2\n    target_label: __address__\n  - action: labelmap\n    regex: __meta_kubernetes_pod_label_(.+)\n  - source_labels: [__meta_kubernetes_namespace]\n    action: replace\n    target_label: namespace\n  - source_labels: [__meta_kubernetes_pod_name]\n    action: replace\n    target_label: pod_name",
>>>>>>> enable istio
      },
    },
    statefulset: base.StatefulSet(metadata, 'prometheus-server') {
      spec+: {
        serviceName: 'prometheus-service',
        replicas: 1,
        template+: {
          spec+: {
            volumes: [
              {
                name: 'prometheus-config-volume',
                configMap: {
                  defaultMode: 420,
                  name: 'prometheus-server-conf',
                },
              },
              {
                name: 'prometheus-datadir',
                persistentVolumeClaim: {
                  claimName: 'prometheus-datadir',
                },
              },
            ],
            initContainers: [
              {
                name: 'init-chown-data-prometheus',
                image: 'busybox:latest',
                volumeMounts: [
                  {
                    name: 'prometheus-datadir',
                    mountPath: '/data/prometheus',
                  },
                ],
                command: [
                  'chown',
                  '-R',
                  '65534:65534',
                  '/data/prometheus'
                ],
              },
            ],
            containers: [
              {
                name: 'prometheus',
                image: 'prom/prometheus',
                args: [
                  '--config.file=/etc/prometheus/prometheus.yml',
                  '--storage.tsdb.path=/data/prometheus/',
                ],
                ports: [
                  {
                    containerPort: 9090,
                  },
                ],
                volumeMounts: [
                  {
                    name: 'prometheus-config-volume',
                    mountPath: '/etc/prometheus/',
                  },
                  {
                    name: 'prometheus-datadir',
                    mountPath: '/data/prometheus/',
                  },
                ],
              },
            ],
          },
        },
        volumeClaimTemplates: [
          {
            metadata: {
              name: 'prometheus-datadir',
            },
            spec: {
              storageClassName: 'standard',
              accessModes: [
                'ReadWriteOnce',
              ],
              resources: {
                requests: {
                  storage: '100Gi',
                },
              },
            },
          },
        ],
      },
    },
    service: base.Service(metadata, 'prometheus-service') {
      app:: 'prometheus-server',
      port:: 9090,
      enable_monitoring:: true,
      spec+: {
        selector: {
          name: 'prometheus-server',
        },
        type: 'NodePort',
        ports: [
          {
            port: 8080,
            targetPort: 9090,
            nodePort: 30000,
          },
        ],
      },
    },
  },
}
