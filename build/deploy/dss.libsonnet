// kube.libsonnet is an import from bitnami, we would not maintain this import this way.
local kube = import "kube.libsonnet"; 
local common = import "common.libsonnet";
local cinit = import "cluster-init-secure.libsonnet";
local crdbExternal = import "cockroachdb-external.libsonnet";
local cockroachdb = import "cockroachdb-statefulset-secure.libsonnet";

{
  // With metadata we can wrap kubectl/kubecfg commands such that they always
  // apply the values in metadata.
  metadata: kube.ConfigMap("cluster-metadata") {
    data: {
      cluster: error "must supply cluster",
      namespace: error "must supply namespace",
    }
  },
  cockroach: {
    shouldInit:: false,
    init: if self.shouldInit then cinit.CockroachInit("asdf") {
      namespace:: "test",
    } else null,
    balanced: crdbExternal.Balanced("10.10.10.10"),
    external: crdbExternal.NodeGateways(["0.0.0.0", "1.1.1.1"]),
    meta: cockroachdb.meta,
    sset: cockroachdb.StatefulSet,
  },
}