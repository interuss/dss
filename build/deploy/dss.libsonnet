local cockroachAuxilliary = import 'cockroachdb-auxilliary.libsonnet';
local cockroachdb = import 'cockroachdb.libsonnet';
local backend = import 'grpc-backend.libsonnet';
local gateway = import 'http-gateway.libsonnet';
local base = import 'base.libsonnet';
local prometheus = import 'prometheus.libsonnet';
local grafana = import 'grafana.libsonnet';
local alertmanager = import 'alertmanager.libsonnet';
# local istio = import 'istio.yaml';
# local certmanager = import 'certmanager/default.yaml';
local certificates = import 'certmanager/config.libsonnet';


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
    # certs: certificates.all(metadata),
    # # certmanager: {
    # #   ["istio-obj-" + i]: certmanager[i],
    # #   for i in std.range(0, std.length(certmanager) - 1)
    # # },
    # external_routing_rule: {
    #   apiVersion: 'networking.istio.io/v1alpha3',
    #   kind: 'DestinationRule',
    #   metadata: {
    #     name: 'tls-foo',
    #   },
    #   spec: {
    #     host: '*.db.mtls-test1.interussplatform.dev',
    #     trafficPolicy: {
    #       tls: {
    #         mode: 'MUTUAL',
    #       },
    #     },
    #   },
    # },
    default_namespace: {
      apiVersion: 'v1',
      kind: 'Namespace',
      metadata: {
        name: metadata.namespace,
        clusterName: metadata.clusterName,
        labels: {
          'istio-injection': 'enabled',
        },
      },
    },
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
    prometheus: prometheus.all(metadata),
    grafana: grafana.all(metadata),
    alertmanager: if metadata.alert.enable == true then alertmanager.all(metadata),
    # istio: if metadata.enable_istio then {
    #   ["istio-obj-" + i]: istio[i],
    #   for i in std.range(0, std.length(istio) - 1)
    # },
  },
}
