{
  all(metadata): {
    cluster_role: {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRole',
      metadata: {
        name: 'kiali',
        labels: {
          app: 'kiali',
          release: 'istio',
        },
      },
      rules: [
        {
          apiGroups: [
            '',
          ],
          resources: [
            'configmaps',
            'endpoints',
            'namespaces',
            'nodes',
            'pods',
            'pods/log',
            'replicationcontrollers',
            'services',
          ],
          verbs: [
            'get',
            'list',
            'watch',
          ],
        },
        {
          apiGroups: [
            'extensions',
            'apps',
          ],
          resources: [
            'deployments',
            'replicasets',
            'statefulsets',
          ],
          verbs: [
            'get',
            'list',
            'watch',
          ],
        },
        {
          apiGroups: [
            'autoscaling',
          ],
          resources: [
            'horizontalpodautoscalers',
          ],
          verbs: [
            'get',
            'list',
            'watch',
          ],
        },
        {
          apiGroups: [
            'batch',
          ],
          resources: [
            'cronjobs',
            'jobs',
          ],
          verbs: [
            'get',
            'list',
            'watch',
          ],
        },
        {
          apiGroups: [
            'config.istio.io',
            'networking.istio.io',
            'authentication.istio.io',
            'rbac.istio.io',
            'security.istio.io',
          ],
          resources: [
            '*',
          ],
          verbs: [
            'create',
            'delete',
            'get',
            'list',
            'patch',
            'watch',
          ],
        },
        {
          apiGroups: [
            'monitoring.kiali.io',
          ],
          resources: [
            'monitoringdashboards',
          ],
          verbs: [
            'get',
            'list',
          ],
        },
      ],
    },
    viewer_cluster_role: {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRole',
      metadata: {
        name: 'kiali-viewer',
        labels: {
          app: 'kiali',
          release: 'istio',
        },
      },
      rules: [
        {
          apiGroups: [
            '',
          ],
          resources: [
            'configmaps',
            'endpoints',
            'namespaces',
            'nodes',
            'pods',
            'pods/log',
            'replicationcontrollers',
            'services',
          ],
          verbs: [
            'get',
            'list',
            'watch',
          ],
        },
        {
          apiGroups: [
            'extensions',
            'apps',
          ],
          resources: [
            'deployments',
            'replicasets',
            'statefulsets',
          ],
          verbs: [
            'get',
            'list',
            'watch',
          ],
        },
        {
          apiGroups: [
            'autoscaling',
          ],
          resources: [
            'horizontalpodautoscalers',
          ],
          verbs: [
            'get',
            'list',
            'watch',
          ],
        },
        {
          apiGroups: [
            'batch',
          ],
          resources: [
            'cronjobs',
            'jobs',
          ],
          verbs: [
            'get',
            'list',
            'watch',
          ],
        },
        {
          apiGroups: [
            'config.istio.io',
            'networking.istio.io',
            'authentication.istio.io',
            'rbac.istio.io',
            'security.istio.io',
          ],
          resources: [
            '*',
          ],
          verbs: [
            'get',
            'list',
            'watch',
          ],
        },
        {
          apiGroups: [
            'monitoring.kiali.io',
          ],
          resources: [
            'monitoringdashboards',
          ],
          verbs: [
            'get',
            'list',
          ],
        },
      ],
    },
    cluster_role_binding: {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRoleBinding',
      metadata: {
        name: 'kiali',
        labels: {
          app: 'kiali',
          release: 'istio',
        },
      },
      roleRef: {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'ClusterRole',
        name: 'kiali',
      },
      subjects: [
        {
          kind: 'ServiceAccount',
          name: 'kiali-service-account',
          namespace: 'istio-system',
        },
      ],
    },
    config_map: {
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: {
        name: 'kiali',
        namespace: 'istio-system',
        labels: {
          app: 'kiali',
          release: 'istio',
        },
      },
      data: {
        'config.yaml': "istio_component_namespaces:\n  grafana: dss-main\n  tracing: istio-system\n  pilot: istio-system\n  prometheus: dss-main\nistio_namespace: istio-system\ndeployment:\n  accessible_namespaces: ['**']\nserver:\n  port: 20001\n  web_root: /kiali\nexternal_services:\n  istio:\n    url_service_version: http://istio-pilot.istio-system:8080/version\n  tracing:\n    url: \n  grafana:\n    url: \n  prometheus:\n    url: http://prometheus.dss-main:9090\n",
      },
    },
    credentials: {
      apiVersion: 'v1',
      kind: 'Secret',
      metadata: {
        name: 'kiali',
        namespace: 'istio-system',
        labels: {
          app: 'kiali',
          release: 'istio',
        },
      },
      type: 'Opaque',
      data: {
        username: 'YWRtaW4=', #admin
        passphrase: 'YWRtaW4=', #admin
      },
    },
    deployment: {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: 'kiali',
        namespace: 'istio-system',
        labels: {
          app: 'kiali',
          release: 'istio',
        },
      },
      spec: {
        replicas: 1,
        selector: {
          matchLabels: {
            app: 'kiali',
          },
        },
        template: {
          metadata: {
            name: 'kiali',
            labels: {
              app: 'kiali',
              release: 'istio',
            },
            annotations: {
              'sidecar.istio.io/inject': 'false',
              'scheduler.alpha.kubernetes.io/critical-pod': '',
              'prometheus.io/scrape': 'true',
              'prometheus.io/port': '9090',
              'kiali.io/runtimes': 'go,kiali',
            },
          },
          spec: {
            serviceAccountName: 'kiali-service-account',
            containers: [
              {
                image: 'quay.io/kiali/kiali:v1.9',
                imagePullPolicy: 'IfNotPresent',
                name: 'kiali',
                command: [
                  '/opt/kiali/kiali',
                  '-config',
                  '/kiali-configuration/config.yaml',
                  '-v',
                  '3',
                ],
                readinessProbe: {
                  httpGet: {
                    path: '/kiali/healthz',
                    port: 20001,
                    scheme: 'HTTP',
                  },
                  initialDelaySeconds: 5,
                  periodSeconds: 30,
                },
                livenessProbe: {
                  httpGet: {
                    path: '/kiali/healthz',
                    port: 20001,
                    scheme: 'HTTP',
                  },
                  initialDelaySeconds: 5,
                  periodSeconds: 30,
                },
                env: [
                  {
                    name: 'ACTIVE_NAMESPACE',
                    valueFrom: {
                      fieldRef: {
                        fieldPath: 'metadata.namespace',
                      },
                    },
                  },
                ],
                volumeMounts: [
                  {
                    name: 'kiali-configuration',
                    mountPath: '/kiali-configuration',
                  },
                  {
                    name: 'kiali-cert',
                    mountPath: '/kiali-cert',
                  },
                  {
                    name: 'kiali-secret',
                    mountPath: '/kiali-secret',
                  },
                ],
                resources: {
                  requests: {
                    cpu: '10m',
                  },
                },
              },
            ],
            volumes: [
              {
                name: 'kiali-configuration',
                configMap: {
                  name: 'kiali',
                },
              },
              {
                name: 'kiali-cert',
                secret: {
                  secretName: 'istio.kiali-service-account',
                  optional: true,
                },
              },
              {
                name: 'kiali-secret',
                secret: {
                  secretName: 'kiali',
                  optional: true,
                },
              },
            ],
            affinity: {
              nodeAffinity: {
                requiredDuringSchedulingIgnoredDuringExecution: {
                  nodeSelectorTerms: [
                    {
                      matchExpressions: [
                        {
                          key: 'beta.kubernetes.io/arch',
                          operator: 'In',
                          values: [
                            'amd64',
                            'ppc64le',
                            's390x',
                          ],
                        },
                      ],
                    },
                  ],
                },
                preferredDuringSchedulingIgnoredDuringExecution: [
                  {
                    weight: 2,
                    preference: {
                      matchExpressions: [
                        {
                          key: 'beta.kubernetes.io/arch',
                          operator: 'In',
                          values: [
                            'amd64',
                          ],
                        },
                      ],
                    },
                  },
                  {
                    weight: 2,
                    preference: {
                      matchExpressions: [
                        {
                          key: 'beta.kubernetes.io/arch',
                          operator: 'In',
                          values: [
                            'ppc64le',
                          ],
                        },
                      ],
                    },
                  },
                  {
                    weight: 2,
                    preference: {
                      matchExpressions: [
                        {
                          key: 'beta.kubernetes.io/arch',
                          operator: 'In',
                          values: [
                            's390x',
                          ],
                        },
                      ],
                    },
                  },
                ],
              },
            },
          },
        },
      },
    },
    service: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: 'kiali',
        namespace: 'istio-system',
        labels: {
          app: 'kiali',
          release: 'istio',
        },
      },
      spec: {
        ports: [
          {
            name: 'http-kiali',
            protocol: 'TCP',
            port: 20001,
          },
        ],
        selector: {
          app: 'kiali',
        },
      },
    },
    service_account: {
      apiVersion: 'v1',
      kind: 'ServiceAccount',
      metadata: {
        name: 'kiali-service-account',
        namespace: 'istio-system',
        labels: {
          app: 'kiali',
          release: 'istio',
        },
      },
    }
  },
}
