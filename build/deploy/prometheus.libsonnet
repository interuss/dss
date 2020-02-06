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
      metadata: {
        name: 'prometheus-server-conf',
        namespace: metadata.namespace,
      },
      data: {
        'prometheus.yml': "global:\n  scrape_interval: 5s\n  evaluation_interval: 5s\n\nscrape_configs:\n\n  - job_name: 'K8s-Endpoints'\n    kubernetes_sd_configs:\n    - role: endpoints\n    relabel_configs:\n    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]\n      action: keep\n      regex: true\n    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]\n      action: replace\n      target_label: __scheme__\n      regex: (https?)\n    - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]\n      action: replace\n      target_label: __metrics_path__\n      regex: (.+)\n    - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]\n      action: replace\n      target_label: __address__\n      regex: ([^:]+)(?::\\d+)?;(\\d+)\n      replacement: $1:$2\n    - action: labelmap\n      regex: __meta_kubernetes_service_label_(.+)\n    - source_labels: [__meta_kubernetes_namespace]\n      action: replace\n      target_label: kubernetes_namespace\n    - source_labels: [__meta_kubernetes_service_name]\n      action: replace\n      target_label: kubernetes_name\n    - source_labels: [__meta_kubernetes_pod_label_statefulset_kubernetes_io_pod_name]\n      action: replace\n      target_label: pod_name\n      regex: (cockroachdb-\\d+)\n    tls_config:\n      insecure_skip_verify: true",
      },
    },
    deployment: base.Deployment(metadata, 'prometheus-deployment') {
      metadata: {
        name: 'prometheus-deployment',
        namespace: metadata.namespace,
      },
      spec: {
        replicas: 1,
        template: {
          metadata: {
            labels: {
              app: 'prometheus-server',
            },
          },
          spec: {
            containers: [
              {
                name: 'prometheus',
                image: 'prom/prometheus',
                args: [
                  '--config.file=/etc/prometheus/prometheus.yml',
                  '--storage.tsdb.path=/prometheus/',
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
                    name: 'prometheus-storage-volume',
                    mountPath: '/prometheus/',
                  },
                ],
              },
            ],
            volumes: [
              {
                name: 'prometheus-config-volume',
                configMap: {
                  defaultMode: 420,
                  name: 'prometheus-server-conf',
                },
              },
              {
                name: 'prometheus-storage-volume',
                emptyDir: {},
              },
            ],
          },
        },
      },
    },
    service: base.Service(metadata, 'prometheus-service') {
      metadata: {
        name: 'prometheus-service',
        namespace: metadata.namespace,
        annotations: {
          'prometheus.io/scrape': 'true',
          'prometheus.io/port': '9090',
        },
      },
      spec: {
        selector: {
          app: 'prometheus-server',
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