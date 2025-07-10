local base = import 'base.libsonnet';
local util = import 'util.libsonnet';
local volumes = import 'volumes.libsonnet';

{
  StatefulSet(metadata): base.StatefulSet(metadata, 'cockroachdb') {
    metadata+: {
      namespace: metadata.namespace,
    },
    spec+: {
      serviceName: 'cockroachdb',
      template+: {
        spec+: {
          serviceAccountName: 'cockroachdb',
          volumes: volumes.cockroachVolumes,
          affinity: {
            podAntiAffinity: {
              preferredDuringSchedulingIgnoredDuringExecution: [
                {
                  weight: 100,
                  podAffinityTerm: {
                    labelSelector: {
                      matchExpressions: [
                        {
                          key: 'app',
                          operator: 'In',
                          values: [
                            'cockroachdb',
                          ],
                        },
                      ],
                    },
                    topologyKey: 'kubernetes.io/hostname',
                  },
                },
              ],
            },
          },
          soloContainer:: base.Container('cockroachdb') {
            image: metadata.cockroach.image,
            volumeMounts: volumes.cockroachMounts,
            ports: [
              {
                name: 'cockroach',
                containerPort: metadata.cockroach.grpc_port,
              },
              {
                name: 'http',
                containerPort: metadata.cockroach.http_port,
              },
            ],
            env: [
              {
                name: 'COCKROACH_CHANNEL',
                value: 'kubernetes-multiregion',
              },
            ],
            livenessProbe: {
              httpGet: {
                path: '/health',
                port: 'http',
                scheme: 'HTTPS',
              },
              initialDelaySeconds: 30,
              periodSeconds: 5,
            },
            readinessProbe: {
              httpGet: {
                path: '/health?ready=1',
                port: 'http',
                scheme: 'HTTPS',
              },
              initialDelaySeconds: 10,
              periodSeconds: 5,
              failureThreshold: 2,
            },
            command: [
              '/bin/bash',
              '-ecx',
              'exec /cockroach/cockroach start ' + std.join(' ', util.makeArgs(self.command_args_)),
            ],
            command_args_:: {
              'certs-dir': '/cockroach/cockroach-certs',
              'advertise-addr': if metadata.single_cluster
                  then '$(hostname -f)'
                  else '${HOSTNAME##*-}.' + metadata.cockroach.hostnameSuffix,
              join: std.join(',', ['cockroachdb-0.cockroachdb'] +
                if metadata.single_cluster then [] else metadata.cockroach.JoinExisting),
              logtostderr: true,
              locality: 'zone=' + metadata.cockroach.locality,
              'locality-advertise-addr': 'zone=' + metadata.cockroach.locality + '@$(hostname -f)',
              'http-addr': '0.0.0.0',
              cache: '25%',
              'max-sql-memory': '25%',
            },
          },
          terminationGracePeriodSeconds: 300,
        },
      },
      podManagementPolicy: 'Parallel',
      volumeClaimTemplates: [
        {
          metadata: {
            name: 'datadir',
          },
          spec: {
            storageClassName: metadata.cockroach.storageClass,
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
}
