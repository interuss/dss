local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';


local googleCockroachLB(metadata, name, ip) = base.Service(metadata, name) {
  port:: metadata.cockroach.grpc_port,
  app:: 'cockroachdb',
  spec+: {
    type: 'LoadBalancer',
    loadBalancerIP: ip,
  },
};

local awsCockroachLB(metadata, name, ip) = base.AWSLoadBalancer(metadata, name, [ip], metadata.subnet) {
  port:: metadata.cockroach.grpc_port,
  app:: 'cockroachdb',
};

local minikubeCockroachLB(metadata, name, ip) = base.Service(metadata, name) {
  port:: metadata.cockroach.grpc_port,
  app:: 'cockroachdb',
};

local cockroachLB(metadata, name, ip) =
    if metadata.cloud_provider == "google" then googleCockroachLB(metadata, name, ip)
    else if metadata.cloud_provider == "aws" then awsCockroachLB(metadata, name, ip)
    else if metadata.cloud_provider == "minikube" then minikubeCockroachLB(metadata, name, ip);
{
  all(metadata): if metadata.datastore == 'cockroachdb' then {
    assert !metadata.cockroach.shouldInit || (metadata.cockroach.shouldInit && metadata.cockroach.JoinExisting == []) : "If shouldInit is True, JoinExisiting should be empty",
    assert metadata.cockroach.locality == "" : "cockroach.locality has been replaced by locality",
    CockroachInit: if metadata.cockroach.shouldInit then base.Job(metadata, 'init') {
      spec+: {
        template+: {
          spec+: {
            volumes_: {
              client_certs: volumes.all(metadata).volumes.client_certs,
              ca_certs: volumes.all(metadata).volumes.ca_certs,
            },
            serviceAccountName: 'cockroachdb',
            soloContainer:: base.Container('cluster-init') {
              image: metadata.cockroach.image,
              command: ['/cockroach/cockroach', 'init'],
              args_:: {
                'certs-dir': '/cockroach/cockroach-certs',
                host: 'cockroachdb-0.cockroachdb.' + metadata.namespace,
              },
              volumeMounts: volumes.all(metadata).mounts.caCert + volumes.all(metadata).mounts.clientCert,
            },
          },
        },
      },
    } else null,

    NodeGateways: if metadata.single_cluster then null else {
      ["gateway-" + i]: cockroachLB(metadata, 'cockroach-db-external-node-' + i, metadata.cockroach.nodeIPs[i]) {
        metadata+: {
          annotations+: {
            'service.alpha.kubernetes.io/tolerate-unready-endpoints': 'true',
          },
        },
        spec+: {
          selector: {
            'statefulset.kubernetes.io/pod-name': 'cockroachdb-' + i,
          },
          publishNotReadyAddresses: true,
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
            name: 'tcp-crdbheadless1',
          },
          {
            port: metadata.cockroach.http_port,
            targetPort: metadata.cockroach.http_port,
            name: 'crdbheadless2',
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
            name: 'tcp-crdbpublic1',
          },
          {
            port: metadata.cockroach.http_port,
            targetPort: metadata.cockroach.http_port,
            name: 'crdbpublic2',
          },
        ],
        sessionAffinity: 'ClientIP',
      },
    },
  } else {}
}
