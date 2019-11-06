local base = import 'base.libsonnet';
<<<<<<< HEAD
local util = import 'util.libsonnet';
=======
>>>>>>> All files in
local volumes = import 'volumes.libsonnet';

{
  StatefulSet(metadata): base.StatefulSet(metadata, 'cockroachdb') {
    metadata+: {
      namespace: metadata.namespace,
    },
    spec+: {
      serviceName: 'cockroachdb',
      replicas: 3,  // default number of replicas.
      template+: {
        metadata+: {
          labels: {
            app: 'cockroachdb',
          },
        },
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
<<<<<<< HEAD
            volumeMounts: volumes.cockroachMounts,
=======
            volumeMounts:: volumes.cockroachMounts,
>>>>>>> All files in
            ports: [
              {
                name: 'cockroach',
                containerPort: metadata.cockroach.grpc_port,
<<<<<<< HEAD
=======
                targetPort: metadata.cockroach.grpc_port,
>>>>>>> All files in
              },
              {
                name: 'http',
                containerPort: metadata.cockroach.http_port,
<<<<<<< HEAD
=======
                targetPort: metadata.cockroach.http_port,
>>>>>>> All files in
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
<<<<<<< HEAD
              'exec /cockroach/cockroach start ' + std.join(' ', util.makeArgs(self.command_args_)),
            ],
            command_args_:: {
              'certs-dir': '/cockroach/cockroach-certs',
              'advertise-addr': '${HOSTNAME##*-}.' + metadata.cockroach.hostnameSuffix,
              join: std.join(',', ['cockroachdb-0.cockroachdb.' + metadata.namespace + '.svc.cluster.local']
                             + metadata.cockroach.JoinExisting),
              logtostderr: true,
              locality: 'zone=' + metadata.cockroach.locality,
=======
              'exec',
              '/cockroach/cockroach',
              'start',
            ],
            args_:: {
              'certs-dir': '/cockroach-certs',
              'advertise-addr': '${HOSTNAME##*-}.' + metadata.cockroach.hostnameSuffix,
              join: 'cockroachdb-0.cockroachdb.' + metadata.namespace + '.svc.cluster.local',
              logtostderr: true,
>>>>>>> All files in
              'locality-advertise-addr': 'zone=' + metadata.cockroach.locality + '@$(hostname -f)',
              'http-addr': '0.0.0.0',
              cache: '25%',
              'max-sql-memory': '25%',
            },
          },
          terminationGracePeriodSeconds: 60,
        },
      },
      podManagementPolicy: 'Parallel',
<<<<<<< HEAD
=======
      updateStrategy: {
        type: 'RollingUpdate',
      },
>>>>>>> All files in
      volumeClaimTemplates: [
        {
          metadata: {
            name: 'datadir',
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
}
