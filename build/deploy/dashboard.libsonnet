local base = import 'base.libsonnet';
local crdbReplicaDash = import 'grafana_dashboards/crdb-replica-grafana.json';
local crdbRuntimeDash = import 'grafana_dashboards/crdb-runtime-grafana.json';
local crdbSqlDash = import 'grafana_dashboards/crdb-sql-grafana.json';
local crdbStorageDash = import 'grafana_dashboards/crdb-storage-grafana.json';
local promOverview = import 'grafana_dashboards/prometheus-overview.json';
local istioCitadel = import 'grafana_dashboards/istio/citadel.json';
local istioGalley = import 'grafana_dashboards/istio/galley.json';
local istioMesh = import 'grafana_dashboards/istio/mesh.json';
local istioMixer = import 'grafana_dashboards/istio/mixer.json';
local istioOverview = import 'grafana_dashboards/istio/overview.json';
local istioPerformance = import 'grafana_dashboards/istio/performance.json';
local istioPilot = import 'grafana_dashboards/istio/pilot.json';
local istioService = import 'grafana_dashboards/istio/service.json';
local istioWorkload = import 'grafana_dashboards/istio/workload.json';
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
      istioDashboards: if metadata.enable_istio then {
        overview: base.ConfigMap(metadata, 'istio-grafana-configuration-dashboards-overview-dashboard') {
          data: {
            'istio-overview.json': std.toString(istioOverview),
          },
        },
        pilot: base.ConfigMap(metadata, 'istio-grafana-configuration-dashboards-pilot-dashboard') {
          data: {
            'pilot-dashboard.json': std.toString(istioPilot),
          },
        },
        mixer: base.ConfigMap(metadata, 'istio-grafana-configuration-dashboards-mixer-dashboard') {
          data: {
            'mixer-dashboard.json': std.toString(istioMixer),
          },
        },
        workload: base.ConfigMap(metadata, 'istio-grafana-configuration-dashboards-workload-dashboard') {
          data: {
            'workload-dashboard.json': std.toString(istioWorkload),
          },
        },
        service: base.ConfigMap(metadata, 'istio-grafana-configuration-dashboards-service-dashboard') {
          data: {
            'service-dashboard.json': std.toString(istioService),
          },
        },
        performance: base.ConfigMap(metadata, 'istio-grafana-configuration-dashboards-performance-dashboard') {
          data: {
            'performance-dashboard.json': std.toString(istioPerformance),
          },
        },
        citadel: base.ConfigMap(metadata, 'istio-grafana-configuration-dashboards-citadel-dashboard') {
          data: {
            'citadel-dashboard.json': std.toString(istioCitadel),
          },
        },
        mesh: base.ConfigMap(metadata, 'istio-grafana-configuration-dashboards-mesh-dashboard') {
          data: {
            'mesh-dashboard.json': std.toString(istioMesh),
          },
        },
        galley: base.ConfigMap(metadata, 'istio-grafana-configuration-dashboards-galley-dashboard') {
          data: {
            'galley-dashboard.json': std.toString(istioGalley),
          },
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
    } + if metadata.enable_istio then {
      grafIstioCitadel: {
        name: 'istio-grafana-configuration-dashboards-citadel-dashboard',
        configMap: {
          defaultMode: 420,
          name: 'istio-grafana-configuration-dashboards-citadel-dashboard',
        },
      },
      grafIstioGalley: {
        name: 'istio-grafana-configuration-dashboards-galley-dashboard',
        configMap: {
          defaultMode: 420,
          name: 'istio-grafana-configuration-dashboards-galley-dashboard',
        },
      },
      grafIstioMesh: {
        name: 'istio-grafana-configuration-dashboards-mesh-dashboard',
        configMap: {
          defaultMode: 420,
          name: 'istio-grafana-configuration-dashboards-mesh-dashboard',
        },
      },
      grafIstioMixer: {
        name: 'istio-grafana-configuration-dashboards-mixer-dashboard',
        configMap: {
          defaultMode: 420,
          name: 'istio-grafana-configuration-dashboards-mixer-dashboard',
        },
      },
      grafIstioOverview: {
        name: 'istio-grafana-configuration-dashboards-pilot-dashboard',
        configMap: {
          defaultMode: 420,
          name: 'istio-grafana-configuration-dashboards-pilot-dashboard',
        },
      },
      grafIstioPerformance: {
        name: 'istio-grafana-configuration-dashboards-performance-dashboard',
        configMap: {
          defaultMode: 420,
          name: 'istio-grafana-configuration-dashboards-performance-dashboard',
        },
      },
      grafIstioPilot: {
        name: 'istio-grafana-configuration-dashboards-pilot-dashboard',
        configMap: {
          defaultMode: 420,
          name: 'istio-grafana-configuration-dashboards-pilot-dashboard',
        },
      },
      grafIstioService: {
        name: 'istio-grafana-configuration-dashboards-service-dashboard',
        configMap: {
          defaultMode: 420,
          name: 'istio-grafana-configuration-dashboards-service-dashboard',
        },
      },
      grafIstioWorkload: {
        name: 'istio-grafana-configuration-dashboards-workload-dashboard',
        configMap: {
          defaultMode: 420,
          name: 'istio-grafana-configuration-dashboards-workload-dashboard',
        },
      },
    } else {},
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
    } + if metadata.enable_istio then {
      grafIstioCitadel: {
        name: 'istio-grafana-configuration-dashboards-citadel-dashboard',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/istio/citadel-dashboard.json',
        subpath: 'citadel-dashbaord.json',
      },
      grafIstioGalley: {
        name: 'istio-grafana-configuration-dashboards-galley-dashboard',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/istio/galley-dashboard.json',
        subpath: 'galley-dashbaord.json',
      },
      grafIstioMesh: {
        name: 'istio-grafana-configuration-dashboards-mesh-dashboard',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/istio/mesh-dashboard.json',
        subpath: 'mesh-dashbaord.json',
      },
      grafIstioMixer: {
        name: 'istio-grafana-configuration-dashboards-mixer-dashboard',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/istio/mixer-dashboard.json',
        subpath: 'mixer-dashbaord.json',
      },
      grafIstioOverview: {
        name: 'istio-grafana-configuration-dashboards-pilot-dashboard',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/istio/pilot-dashboard.json',
        subpath: 'pilot-dashbaord.json',
      },
      grafIstioPerformance: {
        name: 'istio-grafana-configuration-dashboards-performance-dashboard',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/istio/performance-dashboard.json',
        subpath: 'performance-dashbaord.json',
      },
      grafIstioPilot: {
        name: 'istio-grafana-configuration-dashboards-pilot-dashboard',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/istio/pilot-dashboard.json',
        subpath: 'pilot-dashbaord.json',
      },
      grafIstioService: {
        name: 'istio-grafana-configuration-dashboards-service-dashboard',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/istio/service-dashboard.json',
        subpath: 'service-dashbaord.json',
      },
      grafIstioWorkload: {
        name: 'istio-grafana-configuration-dashboards-workload-dashboard',
        readOnly: false,
        mountPath: '/var/lib/grafana/dashboards/istio/workload-dashboard.json',
        subpath: 'workload-dashbaord.json',
      },
    } else {}, 
    mount: util.mapToList(self.mountConfigs),
  },
}