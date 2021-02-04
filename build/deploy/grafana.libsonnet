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

local grafanaConfig = {
  main: {
    app_mode: "production",
  },
  sections: {
    server: {
      domain: "dss-ohio.oneskysystems.com",
      serve_from_sub_path: true,
      root_url: "https://dss-ohio.oneskysystems.com/grafana/"
    },
    security: {
      admin_user: "onesky",
      admin_password: "Dr0ne$"
    },
  }
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

local notifierConfig(metadata) = {
  apiVersion: 1,
  notifiers: [
    {
      name: 'auto-alertmanager',
      type: 'prometheus-alertmanager',
      uid: 'notifier1',
      org_name: 'Main Org.',
      is_default: true,
      send_image: true,
      settings: {
        url: 'http://alertmanager-service.' + metadata.namespace + '.svc:9093',
      },
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
    configMap3: base.ConfigMap(metadata, 'grafana-config') {
      data: {
        'grafana.ini': std.manifestIni(grafanaConfig),
      },
    },
    grafDashboards: dashboard.all(metadata).config,
    notifierConfig: base.ConfigMap(metadata, 'grafana-notifier-provisioning') {
      data: {
        'notifiers.yaml': if metadata.alert.enable == true then std.manifestYamlDoc(notifierConfig(metadata)) else "",
      },
    },

    deployment: base.Deployment(metadata, 'grafana') {
      spec+: {
        replicas: 1,
        selector+: {
          matchLabels+: {
            app: 'grafana',
          },
        },
        template+: {
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
                    mountPath: '/etc/grafana/grafana.ini',
                    subPath: 'grafana.ini',
                    name: 'grafana-config',
                    readOnly: true,
                  },
                  {
                    mountPath: '/etc/grafana/provisioning/notifiers',
                    name: 'grafana-notifier-provisioning',
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
              {
                name: 'grafana-config',
                configMap: {
                  defaultMode: 420,
                  name: 'grafana-config',
                },
              },
              {
                name: 'grafana-notifier-provisioning',
                configMap: {
                  defaultMode: 420,
                  name: 'grafana-notifier-provisioning',
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
