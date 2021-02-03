local base = import 'base.libsonnet'; 
local k8sEndpoints = import 'prometheus_configs/k8s-endpoints.libsonnet';
local istioScrape = import 'prometheus_configs/istio.libsonnet';
local crdbAggregation = import 'prometheus_configs/crdb-aggregation.libsonnet';


local PrometheusConfig(metadata) = {
  global: {
    scrape_interval: '5s',
    evaluation_interval: '5s',
    // label for federated Prometheus.
    external_labels:{
      k8s_cluster: metadata.clusterName,
      environment: metadata.environment,
    },
  },
  rule_files: [
    'aggregation.rules.yml',
  ],
  scrape_configs: k8sEndpoints.scrape_configs + istioScrape.scrape_configs,
};

local PrometheusExternalService(metadata) = base.Service(metadata, 'prometheus-external') {
  app:: 'prometheus',
  port:: 9090,
  spec+: {
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
    configMap: base.ConfigMap(metadata, 'prometheus-conf') {
      data: {
        'prometheus.yml': std.manifestYamlDoc(PrometheusConfig(metadata)),
        'aggregation.rules.yml': std.manifestYamlDoc(crdbAggregation),
      },
    },
    statefulset: base.StatefulSet(metadata, 'prometheus') {
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
                  name: 'prometheus-conf',
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
                  '--storage.tsdb.retention.time=' + metadata.prometheus.retention,
                  // following thanos recommendation
                  '--storage.tsdb.max-block-duration=2h',
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
                  periodSeconds: 6,
                  failureThreshold: 200
                },
                readinessProbe: {
                  httpGet: {
                    path: '/-/ready',
                    port: 9090
                  },
                  initialDelaySeconds: 30,
                  periodSeconds: 6,
                  failureThreshold: 200,
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
              storageClassName: 'gp2',
              accessModes: [
                'ReadWriteOnce',
              ],
              resources: {
                requests: {
                  storage: metadata.prometheus.storage_size,
                },
              },
            },
          },
        ],
      },
    },
    externalService: if metadata.prometheus.expose_external == true then PrometheusExternalService(metadata),
    internalService: base.Service(metadata, 'prometheus-service') {
      app:: 'prometheus',
      port:: 9090,
      enable_monitoring:: true,
    },
  },
}
