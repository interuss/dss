local cockroachAuxiliary = import 'cockroachdb-auxiliary.libsonnet';
local cockroachdb = import 'cockroachdb.libsonnet';
local backend = import 'core-service.libsonnet';
local base = import 'base.libsonnet';
local prometheus = import 'prometheus.libsonnet';
local grafana = import 'grafana.libsonnet';
local alertmanager = import 'alertmanager.libsonnet';
local base = import 'base.libsonnet';
local schema_manager = import 'schema-manager.libsonnet';

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
  all(metadata): {
    default_namespace: base.Namespace(metadata, metadata.namespace) {

    },
    cluster_metadata: base.ConfigMap(metadata, 'cluster-metadata') {
      data: {
        clusterName: metadata.clusterName,
      },
    },
    pspRB: if metadata.PSP.roleBinding then RoleBinding(metadata),

    sset: cockroachdb.StatefulSet(metadata),
    auxiliary: cockroachAuxiliary.all(metadata),
    backend: backend.all(metadata),
    prometheus: prometheus.all(metadata),
    grafana: grafana.all(metadata),
    alertmanager: if metadata.alert.enable == true then alertmanager.all(metadata),
    schema_manager: if metadata.cockroach.shouldInit == true || metadata.schema_manager.enable then schema_manager.all(metadata) else {},
  },
}
