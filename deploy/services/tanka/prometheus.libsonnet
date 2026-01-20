local base = import 'base.libsonnet';
local k8sEndpoints = import 'prometheus_configs/k8s-endpoints.libsonnet';
local crdbAggregation = import 'prometheus_configs/crdb-aggregation.libsonnet';


local PrometheusConfig(metadata) = {
  global: {
    scrape_interval: metadata.prometheus.scrape_interval,
    evaluation_interval: metadata.prometheus.evaluation_interval,
    // label for federated Prometheus.
    external_labels:{
      k8s_cluster: metadata.clusterName,
      environment: metadata.environment,
    },
  },
  rule_files: [
    'aggregation.rules.yml',
    'custom.rules.yml',
  ],
  scrape_configs: k8sEndpoints.scrape_configs,
};

local PrometheusWebConfig(metadata) = {
  tls_server_config: {
    cert_file: '/certs/node.crt',
    key_file: '/certs/node.key',
    client_auth_type: 'RequireAndVerifyClientCert',
    client_ca_file: '/certs/ca.crt'
   }
};

local googleExternalLB(metadata, name, ip) = base.Service(metadata, name) {
  port:: 9090,
  app:: 'prometheus',
  spec+: {
    type: 'LoadBalancer',
    loadBalancerIP: ip,
  },
};

local awsExternalLB(metadata, name, ip) = base.AWSLoadBalancer(metadata, name, [ip], metadata.subnet) {
  port:: 9090,
  app:: 'prometheus',
};

local minikubeExternalLB(metadata, name, ip) = base.Service(metadata, name) {
  port:: 9090,
  app:: 'prometheus',
};

local externalLB(metadata, name, ip) =
    if metadata.cloud_provider == "google" then googleExternalLB(metadata, name, ip)
    else if metadata.cloud_provider == "aws" then awsExternalLB(metadata, name, ip)
    else if metadata.cloud_provider == "minikube" then minikubeExternalLB(metadata, name, ip);

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
        'web-config.yml': std.manifestYamlDoc(PrometheusWebConfig(metadata)),
        'aggregation.rules.yml': std.manifestYamlDoc(crdbAggregation),
        'custom.rules.yml': std.manifestYamlDoc({
          groups: [
            {
              name: 'rules/custom.rules',
              rules: metadata.prometheus.custom_rules
            },
          ],
        }),
      },
    },
    statefulset: base.StatefulSet(metadata, 'prometheus') {
      spec+: {
        serviceName: 'prometheus-service',
        replicas: 1,
        template+: {
          metadata+: {
            annotations+: {
              "checksum/config": std.native('sha256')(std.manifestJson(PrometheusConfig(metadata))),
              "checksum/webconfig": std.native('sha256')(std.manifestJson(PrometheusWebConfig(metadata))),
              "checksum/k8sEndpoints": std.native('sha256')(std.manifestJson(k8sEndpoints)),
              "checksum/crdbAggregation": std.native('sha256')(std.manifestJson(crdbAggregation)),
            },
          },
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
              {
                name: 'prometheus-certs',
                secret: {
                    secretName: 'monitoring.prometheus.certs',
                    defaultMode: 420,
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
                image: metadata.prometheus.image,
                args: [
                  '--config.file=/etc/prometheus/prometheus.yml',
                  '--web.config.file=/etc/prometheus/web-config.yml',
                  '--storage.tsdb.path=/data/prometheus/',
                  '--storage.tsdb.retention.time=' + metadata.prometheus.retention,
                  // following thanos recommendation
                  '--storage.tsdb.max-block-duration=2h',
                ] + metadata.prometheus.custom_args,
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
                  {
                    name: 'prometheus-certs',
                    mountPath: '/certs/',
                  },
                ],
                livenessProbe: {
                  tcpSocket: {
                    port: 9090
                  },
                  initialDelaySeconds: 50,
                  periodSeconds: 6,
                  failureThreshold: 200
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
              storageClassName: metadata.prometheus.storageClass,
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
    externalLB: if metadata.prometheus.expose_external == true then externalLB(metadata, "prometheus", metadata.prometheus.IP),
    internalService: base.Service(metadata, 'prometheus-service') {
      app:: 'prometheus',
      port:: 9090,
    },
  },
}
