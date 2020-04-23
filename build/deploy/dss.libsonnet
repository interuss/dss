local cockroachAuxiliary = import 'cockroachdb-auxiliary.libsonnet';
local cockroachdb = import 'cockroachdb.libsonnet';
local backend = import 'grpc-backend.libsonnet';
local gateway = import 'http-gateway.libsonnet';
local base = import 'base.libsonnet';
local prometheus = import 'prometheus.libsonnet';
local grafana = import 'grafana.libsonnet';
local alertmanager = import 'alertmanager.libsonnet';
local istio = import 'istio/base.libsonnet';
local istio_definitions = import 'istio/custom_resources.libsonnet';
local istio_rules = import 'istio/rules.libsonnet';
local base = import 'base.libsonnet';

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
  all(metadata): if metadata.enable_istio && !metadata.applied_istio_definitions
  then istio_definitions
  else {
    default_namespace: base.Namespace(metadata, metadata.namespace) {
      metadata+: {
        labels+: if metadata.enable_istio then {
          'istio-injection': 'enabled',
        } else {},
      },
    },
    cluster_metadata: base.ConfigMap(metadata, 'cluster-metadata') {
      data: {
        clusterName: metadata.clusterName,
      },
    },
    pspRB: if metadata.PSP.roleBinding then RoleBinding(metadata),

    sset: cockroachdb.StatefulSet(metadata),
    auxiliary: cockroachAuxiliary.all(metadata),
    gateway: gateway.all(metadata),
    backend: backend.all(metadata),
    prometheus: prometheus.all(metadata),
    grafana: grafana.all(metadata),
    alertmanager: if metadata.alert.enable == true then alertmanager.all(metadata),
    istio: if metadata.enable_istio then {
      definitions: istio_definitions,
      base: istio,
      rules: istio_rules.all(metadata),
    },
  },
}
