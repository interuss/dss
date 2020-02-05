local cockroachAuxilliary = import 'cockroachdb-auxilliary.libsonnet';
local cockroachdb = import 'cockroachdb.libsonnet';
local backend = import 'grpc-backend.libsonnet';
local gateway = import 'http-gateway.libsonnet';
local base = import 'base.libsonnet';
local prometheus = import 'prometheus.libsonnet';


local RoleBinding(metadata) = base.RoleBinding(metadata, 'default:privileged') {
  roleRef: {
    apiGroup: 'rbac.authorization.k8s.io',
    kind: 'ClusterRole',
    name: metadata.PSP.roleRef,
  },
  subjects: [
    {
      apiGroup: 'rbac.authorization.k8s.io',
      kind: 'Group',
      name: 'system:serviceaccounts:' + metadata.namespace,
    },
  ],
};

{
  // With metadata we can wrap kubectl/kubecfg commands such that they always
  // apply the values in metadata.
  all(metadata): {
    cluster_metadata: base.ConfigMap(metadata, 'cluster-metadata') {
      data: {
        clusterName: metadata.clusterName,
      },
    },
    pspRB: if metadata.PSP.roleBinding then RoleBinding(metadata) else null,

    sset: cockroachdb.StatefulSet(metadata),
    auxilliary: cockroachAuxilliary.all(metadata),
    gateway: gateway.all(metadata),
    backend: backend.all(metadata),
    prometheus: prometheus.all(metadata)
  },
}
