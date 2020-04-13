local base = import 'base.libsonnet';
local crdbReplicaDash = import 'grafana_dashboards/crdb-replica-grafana.json';
local crdbRuntimeDash = import 'grafana_dashboards/crdb-runtime-grafana.json';
local crdbSqlDash = import 'grafana_dashboards/crdb-sql-grafana.json';
local crdbStorageDash = import 'grafana_dashboards/crdb-storage-grafana.json';
local promOverview = import 'grafana_dashboards/prometheus-overview.json';
local istioOverview = import 'grafana_dashboards/istio-overview.json';
local kubeOverview = import 'grafana_dashboards/kubernetes-overview.json';

local dashboardConfig = {
  apiVersion: 1,
  providers: [
    {
      name: 'default',
      orgId: 1,
      folder: '',
      folderUid: '',
      type: 'file',
      options: {
        path: '/var/lib/grafana/dashboards',
      },
    },
  ],
};

local datasourcePrometheus(metadata) = {
  apiVersion: 1,
  datasources: [
    {
      access: 'proxy',
      editable: true,
      name: 'prometheus',
      orgId: 1,
      type: 'prometheus',
      url: 'http://prometheus-service.' + metadata.namespace + '.svc:9090',
      version: 1,
    },
  ],
};

{
  all(metadata) : {
    configMap: base.ConfigMap(metadata, 'grafana-datasources') {
      data: {
        'prometheus.yaml': std.manifestYamlDoc(datasourcePrometheus(metadata)),
      },
    },
    configMap2: base.ConfigMap(metadata, 'grafana-dash-provisioning') {
      data: {
        'dashboards.yaml': std.manifestYamlDoc(dashboardConfig)
      },
    },
    grafDashConfigMap: base.ConfigMap(metadata, 'grafana.dashboards') {
      data: {
        'crdb-replica-grafana.json': std.toString(crdbReplicaDash),
        'crdb-runtime-grafana.json': std.toString(crdbRuntimeDash),
        'crdb-sql-grafana.json': std.toString(crdbSqlDash),
        'crdb-storage-grafana.json': std.toString(crdbStorageDash),
        'prometheus-overview.json': std.toString(promOverview),
        'kubernetes-overview.json': std.toString(kubeOverview),
        'istio-overview.json': if metadata.enable_istio then std.toString(istioOverview),
      },
    },

    deployment: base.Deployment(metadata, 'grafana') {
      spec: {
        replicas: 1,
        selector: {
          matchLabels: {
            app: 'grafana',
          },
        },
        template: {
          metadata: {
            name: 'grafana',
            labels: {
              app: 'grafana',
            },
          },
          spec: {
            containers: [
              {
                name: 'grafana',
                image: 'grafana/grafana:latest',
                ports: [
                  {
                    name: 'grafana',
                    containerPort: 3000,
                  },
                ],
                resources: {
                  limits: {
                    memory: '2Gi',
                    cpu: '1000m',
                  },
                  requests: {
                    memory: '1Gi',
                    cpu: '500m',
                  },
                },
                volumeMounts: [
                  {
                    mountPath: '/var/lib/grafana',
                    name: 'grafana-storage',
                  },
                  {
                    mountPath: '/etc/grafana/provisioning/datasources',
                    name: 'grafana-datasources',
                    readOnly: false,
                  },
                  {
                    mountPath: '/etc/grafana/provisioning/dashboards',
                    name: 'grafana-dash-provisioning',
                    readOnly: false,
                  },
                  {
                    mountPath: '/var/lib/grafana/dashboards',
                    name: 'grafana-dashboards-json',
                    readOnly: false,
                  },
                ],
              },
            ],
            volumes: [
              {
                name: 'grafana-storage',
                emptyDir: {},
              },
              {
                name: 'grafana-datasources',
                configMap: {
                  defaultMode: 420,
                  name: 'grafana-datasources',
                },
              },
              {
                name: 'grafana-dash-provisioning',
                configMap: {
                  defaultMode: 420,
                  name: 'grafana-dash-provisioning',
                },
              },
              {
                name: 'grafana-dashboards-json',
                configMap: {
                  defaultMode: 420,
                  name: 'grafana.dashboards',
                },
              },
            ],
          },
        },
      },
    },
    service: base.Service(metadata, 'grafana') {
      app:: 'grafana',
      port:: 3000,
      enable_monitoring:: true,
    },
  },
}
