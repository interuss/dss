local base = import 'base.libsonnet'; 

local PrometheusExternalService(metadata) = base.Service(metadata, 'prometheus-external') {
  app:: 'prometheus-server',
  port:: 9090,
  spec+: {
    selector: {
      name: 'prometheus-server',
    },
    type: 'LoadBalancer',
    loadBalancerIP: metadata.prometheus.IP,
    loadBalancerSourceRanges: metadata.prometheus.whitelist_ip_ranges
  }
};

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
        'prometheus.yml': "global:\n  external_labels:\n    cluster_name: " + metadata.clusterName + "\n  scrape_interval: 5s\n  evaluation_interval: 5s\n\nscrape_configs:\n\n  - job_name: 'K8s-Endpoints'\n    kubernetes_sd_configs:\n    - role: endpoints\n    relabel_configs:\n    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]\n      action: keep\n      regex: true\n    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]\n      action: replace\n      target_label: __scheme__\n      regex: (https?)\n    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]\n      action: replace\n      target_label: __metrics_path__\n      regex: (.+)\n    - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]\n      action: replace\n      target_label: __address__\n      regex: ([^:]+)(?::\\d+)?;(\\d+)\n      replacement: $1:$2\n    - action: labelmap\n      regex: __meta_kubernetes_service_label_(.+)\n    - source_labels: [__meta_kubernetes_namespace]\n      action: replace\n      target_label: kubernetes_namespace\n    - source_labels: [__meta_kubernetes_service_name]\n      action: replace\n      target_label: kubernetes_name\n    - source_labels: [__meta_kubernetes_pod_label_statefulset_kubernetes_io_pod_name]\n      action: replace\n      target_label: pod_name\n      regex: (cockroachdb-\\d+)\n    tls_config:\n      insecure_skip_verify: true",
        'aggregation.rules.yml': 'groups:\n- name: rules/aggregation.rules\n  rules:\n  - record: node:capacity\n    expr: sum without(store) (capacity{app="cockroachdb"})\n  - record: cluster:capacity\n    expr: sum without(instance) (node:capacity{app="cockroachdb"})\n  - record: node:capacity_available\n    expr: sum without(store) (capacity_available{app="cockroachdb"})\n  - record: cluster:capacity_available\n    expr: sum without(instance) (node:capacity_available{app="cockroachdb"})\n  - record: capacity_available:ratio\n    expr: capacity_available{app="cockroachdb"} / capacity{app="cockroachdb"}\n  - record: node:capacity_available:ratio\n    expr: node:capacity_available{app="cockroachdb"} / node:capacity{app="cockroachdb"}\n  - record: cluster:capacity_available:ratio\n    expr: cluster:capacity_available{app="cockroachdb"} / cluster:capacity{app="cockroachdb"}\n  # Histogram rules: these are fairly expensive to compute live, so we precompute a few percetiles.\n  - record: txn_durations_bucket:rate1m\n    expr: rate(txn_durations_bucket{app="cockroachdb"}[1m])\n  - record: txn_durations:rate1m:quantile_50\n    expr: histogram_quantile(0.5, txn_durations_bucket:rate1m)\n  - record: txn_durations:rate1m:quantile_75\n    expr: histogram_quantile(0.75, txn_durations_bucket:rate1m)\n  - record: txn_durations:rate1m:quantile_90\n    expr: histogram_quantile(0.9, txn_durations_bucket:rate1m)\n  - record: txn_durations:rate1m:quantile_95\n    expr: histogram_quantile(0.95, txn_durations_bucket:rate1m)\n  - record: txn_durations:rate1m:quantile_99\n    expr: histogram_quantile(0.99, txn_durations_bucket:rate1m)\n  - record: exec_latency_bucket:rate1m\n    expr: rate(exec_latency_bucket{app="cockroachdb"}[1m])\n  - record: exec_latency:rate1m:quantile_50\n    expr: histogram_quantile(0.5, exec_latency_bucket:rate1m)\n  - record: exec_latency:rate1m:quantile_75\n    expr: histogram_quantile(0.75, exec_latency_bucket:rate1m)\n  - record: exec_latency:rate1m:quantile_90\n    expr: histogram_quantile(0.9, exec_latency_bucket:rate1m)\n  - record: exec_latency:rate1m:quantile_95\n    expr: histogram_quantile(0.95, exec_latency_bucket:rate1m)\n  - record: exec_latency:rate1m:quantile_99\n    expr: histogram_quantile(0.99, exec_latency_bucket:rate1m)\n  - record: round_trip_latency_bucket:rate1m\n    expr: rate(round_trip_latency_bucket{app="cockroachdb"}[1m])\n  - record: round_trip_latency:rate1m:quantile_50\n    expr: histogram_quantile(0.5, round_trip_latency_bucket:rate1m)\n  - record: round_trip_latency:rate1m:quantile_75\n    expr: histogram_quantile(0.75, round_trip_latency_bucket:rate1m)\n  - record: round_trip_latency:rate1m:quantile_90\n    expr: histogram_quantile(0.9, round_trip_latency_bucket:rate1m)\n  - record: round_trip_latency:rate1m:quantile_95\n    expr: histogram_quantile(0.95, round_trip_latency_bucket:rate1m)\n  - record: round_trip_latency:rate1m:quantile_99\n    expr: histogram_quantile(0.99, round_trip_latency_bucket:rate1m)\n  - record: sql_exec_latency_bucket:rate1m\n    expr: rate(sql_exec_latency_bucket{app="cockroachdb"}[1m])\n  - record: sql_exec_latency:rate1m:quantile_50\n    expr: histogram_quantile(0.5, sql_exec_latency_bucket:rate1m)\n  - record: sql_exec_latency:rate1m:quantile_75\n    expr: histogram_quantile(0.75, sql_exec_latency_bucket:rate1m)\n  - record: sql_exec_latency:rate1m:quantile_90\n    expr: histogram_quantile(0.9, sql_exec_latency_bucket:rate1m)\n  - record: sql_exec_latency:rate1m:quantile_95\n    expr: histogram_quantile(0.95, sql_exec_latency_bucket:rate1m)\n  - record: sql_exec_latency:rate1m:quantile_99\n    expr: histogram_quantile(0.99, sql_exec_latency_bucket:rate1m)\n  - record: raft_process_logcommit_latency_bucket:rate1m\n    expr: rate(raft_process_logcommit_latency_bucket{app="cockroachdb"}[1m])\n  - record: raft_process_logcommit_latency:rate1m:quantile_50\n    expr: histogram_quantile(0.5, raft_process_logcommit_latency_bucket:rate1m)\n  - record: raft_process_logcommit_latency:rate1m:quantile_75\n    expr: histogram_quantile(0.75, raft_process_logcommit_latency_bucket:rate1m)\n  - record: raft_process_logcommit_latency:rate1m:quantile_90\n    expr: histogram_quantile(0.9, raft_process_logcommit_latency_bucket:rate1m)\n  - record: raft_process_logcommit_latency:rate1m:quantile_95\n    expr: histogram_quantile(0.95, raft_process_logcommit_latency_bucket:rate1m)\n  - record: raft_process_logcommit_latency:rate1m:quantile_99\n    expr: histogram_quantile(0.99, raft_process_logcommit_latency_bucket:rate1m)\n  - record: raft_process_commandcommit_latency_bucket:rate1m\n    expr: rate(raft_process_commandcommit_latency_bucket{app="cockroachdb"}[1m])\n  - record: raft_process_commandcommit_latency:rate1m:quantile_50\n    expr: histogram_quantile(0.5, raft_process_commandcommit_latency_bucket:rate1m)\n  - record: raft_process_commandcommit_latency:rate1m:quantile_75\n    expr: histogram_quantile(0.75, raft_process_commandcommit_latency_bucket:rate1m)\n  - record: raft_process_commandcommit_latency:rate1m:quantile_90\n    expr: histogram_quantile(0.9, raft_process_commandcommit_latency_bucket:rate1m)\n  - record: raft_process_commandcommit_latency:rate1m:quantile_95\n    expr: histogram_quantile(0.95, raft_process_commandcommit_latency_bucket:rate1m)\n  - record: raft_process_commandcommit_latency:rate1m:quantile_99\n    expr: histogram_quantile(0.99, raft_process_commandcommit_latency_bucket:rate1m)\n'
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
                livenessProbe: {
                  httpGet: {
                    path: '/-/healthy',
                    port: 9090
                  },
                  initialDelaySeconds: 50,
                  periodSeconds: 5,
                },
                readinessProbe: {
                  httpGet: {
                    path: '/-/ready',
                    port: 9090
                  },
                  initialDelaySeconds: 30,
                  periodSeconds: 5,
                  failureThreshold: 5,
                },
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
    externalService: if metadata.prometheus.external == true then PrometheusExternalService(metadata),
    internalService: base.Service(metadata, 'prometheus-service') {
      app:: 'prometheus-server',
      port:: 9090,
      enable_monitoring:: true,
      spec+: {
        selector: {
          name: 'prometheus-server',
        },
        type: 'ClusterIP',
      },
    },
  },
}
