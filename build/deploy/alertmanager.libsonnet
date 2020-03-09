local base = import 'base.libsonnet'; 

{
  all(metadata): {
    configMap: base.ConfigMap(metadata, 'alertmanager-config') {
      data: {
        'alertmanager.yaml' : 'global:\n  smtp_smarthost: \'' + metadata.alert.smtp.host + '\'\n  smtp_from: \'' + metadata.alert.smtp.email + '\'\n  smtp_auth_username: \'' + metadata.alert.smtp.email + '\'\n  smtp_auth_password: "' + metadata.alert.smtp.password + '"\n\n\ntemplates:\n- \'/etc/alertmanager/template/*.tmpl\'\n\nroute:\n  group_by: [\'alertname\']\n\n  group_wait: 30s\n\n  group_interval: 5m\n\n  repeat_interval: 3h\n\n  receiver: dss-team\n\nreceivers:\n- name: \'dss-team\'\n  email_configs:\n  - to: \'' + metadata.alert.smtp.dest + '\''
      },
    },
    deployment: base.Deployment(metadata, 'alertmanager-deployment') {
      app:: 'alertmanager-deployment',
      metadata+: {
        namespace: metadata.namespace,
      },
      spec+: {
				replicas: 1,
        template+: {
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