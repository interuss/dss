local base = import 'base.libsonnet'; 
local k8sEndpoints = import 'prometheus-configs/k8s-endpoints.libsonnet';
local istioScrape = import 'prometheus-configs/istio.libsonnet';
local crdbAggregation = import 'prometheus-configs/crdb-aggregation.libsonnet';


local prometheusConfig = {
  global: {
    scrape_interval: '5s',
    evaluation_interval: '5s',
  },
  rule_files: [
    'aggregation.rules.yml',
  ],
  scrape_configs: k8sEndpoints.scrape_configs + istioScrape.scrape_configs,
};

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
        'prometheus.yml': std.manifestYamlDoc(prometheusConfig),
        'aggregation.rules.yml': std.manifestYamlDoc(crdbAggregation),
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
    externalService: if metadata.prometheus.expose_external == true then PrometheusExternalService(metadata),
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
