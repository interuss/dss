// kube.libsonnet is an import from bitnami, we would not maintain this import this way.
local kube = import "kube.libsonnet"; 
local common = import "common.libsonnet";

local crLabels = {
  metadata+: {
    labels: {
      app: "cockroachdb",
    }
  }
};

{
  meta: {
    svcAccount: kube.ServiceAccount("cockroachdb") + crLabels,
  },

  StatefulSet: kube.StatefulSet("cockroachdb") + crLabels {
    local dbHostnameSuffix_ = self.dbHostnameSuffix,
    local locality_ = self.locality,
    local namespace_ = self.namespace,
    locality:: error "must supply locality",
    namespace: "default-ns",   # TODO: set namespace better.
    spec+: {
      serviceName: "cockroachdb",
      replicas: 3,  # default number of replicas.
      template+: crLabels + {
        spec+: {
          serviceAccountName: 'cockroachdb',
          volumes: common.cockroach.volumes,
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
          containers_:: {
            cockroachdb: kube.Container("cockroachdb") {
              // TODO stub this.
              image: "cockroachdb",
              volumeMounts:: common.cockroach.volumeMounts,
              ports: [
                {
                  containerPort: common.cockroach.grpc_port,
                  targetPort: common.cockroach.http_port,
                },
              ],
              env: [
                {
                  name: 'COCKROACH_CHANNEL',
                  value: 'kubernetes-multiregion'
                },
              ],
              livenessProbe: {
                httpGet: {
                  path: "/health",
                  port: "http",
                  scheme: "HTTPS",
                },
                initialDelaySeconds: 30,
                periodSeconds: 5,
              },
              readinessProbe: {
                httpGet: {
                  path: "/health?ready=1",
                  port: "http",
                  scheme: "HTTPS",
                },
                initialDelaySeconds: 10,
                periodSeconds: 5,
                failureThreshold: 2
              },
              command: [
                "/bin/bash",
                "-ecx",
                "exec",
                "/cockroach/cockroach",
                "start",
              ],
              args_:: {
                "certs-dir": "/cockroach-certs",
                "advertise-addr": "${HOSTNAME##*-}." + dbHostnameSuffix_,
                join: "cockroachdb-0.cockroachdb." + namespace_ + ".svc.cluster.local",
                logtostderr: true,
                "locality-advertise-addr": "zone=" + locality_ +"@$(hostname -f)",
                "http-addr": "0.0.0.0",
                cache: "25%",
                "max-sql-memory": "25%",
              },
            },
          },
          terminationGracePeriodSeconds: 60,
        },
      },
      podManagementPolicy: 'Parallel',
      updateStrategy: {
        type: 'RollingUpdate',
      },
      volumeClaimTemplates: [
        {
          metadata: {
            name: 'datadir',
          },
          spec: {
            storageClassName: "standard",
            accessModes: [
              'ReadWriteOnce',
            ],
            resources: {
              requests: {
                storage: "100Gi",
              },
            },
          },
        },
      ],
    },
  },
}