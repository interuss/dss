local base = import 'base.libsonnet';
local crdbReplicaDash = import 'grafana_dashboards/crdb-replica-grafana.json';
local crdbRuntimeDash = import 'grafana_dashboards/crdb-runtime-grafana.json';
local crdbSqlDash = import 'grafana_dashboards/crdb-sql-grafana.json';
local crdbStorageDash = import 'grafana_dashboards/crdb-storage-grafana.json';
local promOverview = import 'grafana_dashboards/prometheus-overview.json';
local kubeOverview = import 'grafana_dashboards/kubernetes-overview.json';
local util = import 'util.libsonnet';
{
    all(metadata): {
    config: {
      grafCrdbReplica: base.ConfigMap(metadata, 'grafana-crdb-replica') {
        data: {
          'crdb-replica-grafana.json': std.toString(crdbReplicaDash),
        },
      },
      grafCrdbRuntime: base.ConfigMap(metadata, 'grafana-crdb-runtime') {
        data: {
          'crdb-runtime-grafana.json': std.toString(crdbRuntimeDash),
        },
      },
      grafCrdbSql: base.ConfigMap(metadata, 'grafana-crdb-sql') {
        data: {
          'crdb-sql-grafana.json': std.toString(crdbSqlDash),
        },
      },
      grafCrdbStorage: base.ConfigMap(metadata, 'grafana-crdb-storage') {
        data: {
          'crdb-storage-grafana.json': std.toString(crdbStorageDash),
        },
      },
      grafPromOverview: base.ConfigMap(metadata, 'grafana-prometheus-overview') {
        data: {
          'prometheus-overview.json': std.toString(promOverview),
        },
      },
      grafKubeOverview: base.ConfigMap(metadata, 'grafana-kube-overview') {
        data: {
          'kubernetes-overview.json': std.toString(kubeOverview),
        },
      },
    },
    volumeConfigs: {
      grafCrdbReplica: {
        name: 'grafana-crdb-replica',
        configMap: {
          defaultMode: 420,
          name: 'grafana-crdb-replica',
        },
      },
      grafCrdbRuntime: {
        name: 'grafana-crdb-runtime',
        configMap: {
          defaultMode: 420,
          name: 'grafana-crdb-runtime',
        },
      },
      grafCrdbSql: {
        name: 'grafana-crdb-sql',
        configMap: {
          defaultMode: 420,
          name: 'grafana-crdb-sql',
        },
      },
      grafCrdbStorage: {
        name: 'grafana-crdb-storage',
        configMap: {
          defaultMode: 420,
          name: 'grafana-crdb-storage',
        },
      },
      grafPromOverview: {
        name: 'grafana-prometheus-overview',
        configMap: {
          defaultMode: 420,
          name: 'grafana-prometheus-overview',
        },
      },
      grafKubeOverview: {
        name: 'grafana-kube-overview',
        configMap: {
          defaultMode: 420,
          name: 'grafana-kube-overview',
        },
      },
    },
    volumes: util.mapToList(self.volumeConfigs),
    mountConfigs: {
      grafCrdbReplica: {
        name: 'grafana-crdb-replica',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/grafana-crdb-replica',
      },
      grafCrdbRuntime: {
        name: 'grafana-crdb-runtime',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/grafana-crdb-runtime',
      },
      grafCrdbSql: {
        name: 'grafana-crdb-sql',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/grafana-crdb-sql',
      },
      grafCrdbStorage: {
        name: 'grafana-crdb-storage',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/grafana-crdb-storage',
      },
      grafPromOverview: {
        name: 'grafana-prometheus-overview',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/grafana-prometheus-overview',
      },
      grafKubeOverview: {
        name: 'grafana-kube-overview',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/grafana-kube-overview',
      },
    },
    mount: util.mapToList(self.mountConfigs),
  },
}
