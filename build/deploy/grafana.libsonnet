local base = import 'base.libsonnet';
local dashboard = import 'dashboard.libsonnet';

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
      allowUiUpdates: true,
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
    grafDashboards: dashboard.all(metadata).config,

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
                ] + dashboard.all(metadata).mount,
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
              
            ] + dashboard.all(metadata).volumes,
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
