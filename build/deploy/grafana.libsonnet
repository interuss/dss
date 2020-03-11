local base = import 'base.libsonnet';

{
  all(metadata) : {
    configMap: base.ConfigMap(metadata, 'grafana-datasources') {
      data: {
        'prometheus.yaml': '{\n    "apiVersion": 1,\n    "datasources": [\n        {\n           "access":"proxy",\n            "editable": true,\n            "name": "prometheus",\n            "orgId": 1,\n            "type": "prometheus",\n            "url": "http://prometheus-service.' + metadata.namespace + '.svc:8080",\n            "version": 1\n        }\n    ]\n}',
      },
    },
    configMap2: base.ConfigMap(metadata, 'grafana-dash-provisioning') {
      data: {
        'dashboards.yaml': "apiVersion: 1\n\nproviders:\n - name: 'default'\n   orgId: 1\n   folder: ''\n   folderUid: ''\n   type: file\n   options:\n     path: /var/lib/grafana/dashboards\n",
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
