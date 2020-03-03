local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';

{
  all(metadata): {
    CockroachInit: if metadata.cockroach.shouldInit then base.Job(metadata, 'init') {
      spec+: {
        template+: {
          spec+: {
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

    NodeGateways: {
      ["gateway-" + i]: cockroachLB(metadata, 'cockroach-db-external-node-' + i, metadata.cockroach.nodeIPs[i]) {
        spec+: {
          selector: {
            'statefulset.kubernetes.io/pod-name': 'cockroachdb-' + i,
          },
        },
      }
      for i in std.range(0, std.length(metadata.cockroach.nodeIPs) - 1)
    },

    svcAccount: base.ServiceAccount(metadata, 'cockroachdb') {
      metadata+: {
        labels: {
          app: 'cockroachdb',
        },
      },
    },

    role: base.Role(metadata, 'cockroachdb') {
      app:: 'cockroachdb',
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


    clusterRole: base.ClusterRole(metadata, 'cockroachdb') {
      app:: 'cockroachdb',
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

    roleBinding: base.RoleBinding(metadata, 'cockroachdb') {
      app:: 'cockroachdb',
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

    clusterRoleBinding: base.ClusterRoleBinding(metadata, 'cockroachdb') {
      app:: 'cockroachdb',
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

    service: base.Service(metadata, 'cockroachdb') {
      app:: 'cockroachdb',
      enable_monitoring:: true,
      metadata+: {
        annotations+: {
          'service.alpha.kubernetes.io/tolerate-unready-endpoints': 'true',
          'prometheus.io/port': std.toString(metadata.cockroach.http_port),
          'prometheus.io/scheme': 'https',
          'prometheus.io/path': '/_status/vars',
        },
      },
      spec+: {
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
      },
    },

    podDisruptionBudget: base.PodDisruptionBudget(metadata, 'cockroachdb-budget') {
      app:: 'cockroachdb',
    },

    cockroachBalanced: base.Service(metadata, 'cockroachdb-balanced') {
      app:: 'cockroachdb',
      spec+: {
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
        sessionAffinity: 'ClientIP',
      },
    },
  },
}
