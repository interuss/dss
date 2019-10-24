// kube.libsonnet is an import from bitnami, we would not maintain this import this way.
local kube = import "kube.libsonnet"; 
local common = import "common.libsonnet";

{
  CockroachInit(target_pod): kube.Job("frontend") {
    local job = self,
    cockroach_port:: 26257,
    namespace:: error "must specify namespace",
    spec+: {
      template+: {
        spec+: {
          volumes: [
            common.cockroach.volumes.client_certs,
          ],
          serviceAccountName: "cockroachdb",
          containers_:: {
            cluster_init: kube.Container("cluster-init") {
              // TODO stub this.
              image: "cockroachdb",
              command: ["/cockroach/cockroach", "init"],
              args_:: {
                "certs-dir": "/cockroach-certs",
                "host": "cockroachdb-0.cockroachdb." + job.namespace + ".svc.cluster.local:"+ job.cockroach_port,
              },
              volumeMounts:: [
                common.volumeMounts.client_certs,
              ],
            },
          },
        },
      },
    },
  },
}