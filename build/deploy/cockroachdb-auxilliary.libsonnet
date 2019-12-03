local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';

local cockroachLB(metadata, name, ip) = base.Service(metadata, name) {
  metadata+: {
    namespace: metadata.namespace,
  },
  port:: metadata.cockroach.grpc_port,
  app: 'cockroachdb',
  spec+: {
    type: 'LoadBalancer',
    loadBalancerIP: ip,
  },
};

{
  all(metadata): {
    CockroachInit: if metadata.cockroach.shouldInit then base.Job(metadata, 'init') {
      spec+: {
        template+: {
          spec+: {
            volumes_: {
              client_certs: volumes.volumes.client_certs,
              ca_certs: volumes.volumes.ca_certs,
            },
            serviceAccountName: 'cockroachdb',
            soloContainer:: base.Container('cluster-init') {
              image: metadata.cockroach.image,
              command: ['/cockroach/cockroach', 'init'],
              args_:: {
                'certs-dir': '/cockroach/cockroach-certs',
                host: 'cockroachdb-0.cockroachdb.' + metadata.namespace + '.svc.cluster.local:' + metadata.cockroach.grpc_port,
              },
              volumeMounts: volumes.mounts.caCert + volumes.mounts.clientCert,
            },
          },
        },
      },
    } else null,

    Balanced: cockroachLB(metadata, 'cockroach-db-external-balanced', metadata.cockroach.balancedIP),

    NodeGateways: [
      cockroachLB(metadata, 'cockroach-db-external-node-' + i, metadata.cockroach.nodeIPs[i]) {
        spec+: {
          selector: {
            'statefulset.kubernetes.io/pod-name': 'cockroachdb-' + i,
          },
        },
      }
      for i in std.range(0, std.length(metadata.cockroach.nodeIPs) - 1)
    ],

    svcAccount: {
      apiVersion: 'v1',
      kind: 'ServiceAccount',
      metadata: {
        namespace: metadata.namespace,
        name: 'cockroachdb',
        labels: {
          app: 'cockroachdb',
        },
      },
    },

    role: {
      apiVersion: 'rbac.authorization.k8s.io/v1beta1',
      kind: 'Role',
      metadata: {
        namespace: metadata.namespace,
        name: 'cockroachdb',
        labels: {
          app: 'cockroachdb',
        },
      },
      rules: [
        {
          apiGroups: [
            '',
          ],
          resources: [
            'secrets',
          ],
          verbs: [
            'create',
            'get',
          ],
        },
      ],
    },


    clusterRole: {
      apiVersion: 'rbac.authorization.k8s.io/v1beta1',
      kind: 'ClusterRole',
      metadata: {
        name: 'cockroachdb',
        labels: {
          app: 'cockroachdb',
        },
      },
      rules: [
        {
          apiGroups: [
            'certificates.k8s.io',
          ],
          resources: [
            'certificatesigningrequests',
          ],
          verbs: [
            'create',
            'get',
            'watch',
          ],
        },
      ],
    },

    roleBinding: {
      apiVersion: 'rbac.authorization.k8s.io/v1beta1',
      kind: 'RoleBinding',
      metadata: {
        namespace: metadata.namespace,
        name: 'cockroachdb',
        labels: {
          app: 'cockroachdb',
        },
      },
      roleRef: {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'Role',
        name: 'cockroachdb',
      },
      subjects: [
        {
          kind: 'ServiceAccount',
          name: 'cockroachdb',
          namespace: metadata.namespace,
        },
      ],
    },

    clusterRoleBinding: {
      apiVersion: 'rbac.authorization.k8s.io/v1beta1',
      kind: 'ClusterRoleBinding',
      metadata: {
        name: 'cockroachdb',
        labels: {
          app: 'cockroachdb',
        },
      },
      roleRef: {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'ClusterRole',
        name: 'cockroachdb',
      },
      subjects: [
        {
          kind: 'ServiceAccount',
          name: 'cockroachdb',
          namespace: 'default',
        },
      ],
    },

    service: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        namespace: metadata.namespace,
        name: 'cockroachdb',
        labels: {
          app: 'cockroachdb',
        },
        annotations: {
          'service.alpha.kubernetes.io/tolerate-unready-endpoints': 'true',
          'prometheus.io/scrape': 'true',
          'prometheus.io/path': '_status/vars',
          'prometheus.io/port': std.toString(metadata.cockroach.http_port),
        },
      },
      spec: {
        ports: [
          {
            port: metadata.cockroach.grpc_port,
            targetPort: metadata.cockroach.grpc_port,
            name: 'cockroach',
          },
          {
            port: metadata.cockroach.http_port,
            targetPort: metadata.cockroach.http_port,
            name: 'http',
          },
        ],
        publishNotReadyAddresses: true,
        clusterIP: 'None',
        selector: {
          app: 'cockroachdb',
        },
      },
    },

    podDisruptionBudget: {
      apiVersion: 'policy/v1beta1',
      kind: 'PodDisruptionBudget',
      metadata: {
        namespace: metadata.namespace,
        name: 'cockroachdb-budget',
        labels: {
          app: 'cockroachdb',
        },
      },
      spec: {
        selector: {
          matchLabels: {
            app: 'cockroachdb',
          },
        },
        maxUnavailable: 1,
      },
    },

    cockroachBalanced: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        namespace: metadata.namespace,
        name: 'cockroachdb-balanced',
        labels: {
          app: 'cockroachdb',
        },
      },
      spec: {
        ports: [
          {
            port: metadata.cockroach.grpc_port,
            targetPort: metadata.cockroach.grpc_port,
            name: 'cockroach',
          },
          {
            port: metadata.cockroach.http_port,
            targetPort: metadata.cockroach.http_port,
            name: 'http',
          },
        ],
        selector: {
          app: 'cockroachdb',
        },
        sessionAffinity: 'ClientIP',
      },
    },
  },
}
