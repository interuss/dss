local base = import 'base.libsonnet'; 

local alertmanagerConfig(metadata) = {
  global: {
    smtp_smarthost: metadata.alert.smtp.host,
    smtp_from: metadata.alert.smtp.email,
    smtp_auth_username: metadata.alert.smtp.email,
    smtp_auth_password: metadata.alert.smtp.password,
  },
  templates: [
    '/etc/alertmanager/template/*.tmpl',
  ],
  route: {
    group_by: [
      'alertname',
    ],
    group_wait: '30s',
    group_interval: '5m',
    repeat_interval: '3h',
    receiver: 'dss-team',
  },
  receivers: [
    {
      name: 'dss-team',
      email_configs: [
        {
          to: metadata.alert.smtp.dest,
        },
      ],
    },
  ],
};

{
  all(metadata): {
    configMap: base.ConfigMap(metadata, 'alertmanager-config') {
      data: {
        'alertmanager.yaml' : std.manifestYamlDoc(alertmanagerConfig(metadata)),
      },
    },
    deployment: base.Deployment(metadata, 'alertmanager-deployment') {
      metadata+: {
        namespace: metadata.namespace,
      },
      spec+: {
				replicas: 1,
        template+: {
          metadata+: {
            labels+: {
              app: 'alertmanager',
            },
          },
          spec+: {
            containers: [
              {
                name: 'alertmanager',
                image: 'prom/alertmanager',
                args: [
                  '--config.file=/etc/alertmanager/alertmanager.yaml',
                ],
                volumeMounts: [
                  {
                    name: 'alertmanager-config',
                    mountPath: '/etc/alertmanager/alertmanager.yaml',
                    subPath: 'alertmanager.yaml',
                  },
                ],
                ports: [
                  {
                    name: 'alertmanager',
                    containerPort: 9093,
                  },
                ],
              },
            ],
            volumes: [
              {
                name: 'alertmanager-config',
                configMap: {
                  defaultMode: 420,
                  name: 'alertmanager-config',
                },
              },
            ],
          },
        },
      },
    },
    service: base.Service(metadata, 'alertmanager-service') {
      app:: 'alertmanager-deployment',
      port:: 9093,
    },
  },
}