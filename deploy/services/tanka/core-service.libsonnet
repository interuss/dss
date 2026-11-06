local base = import 'base.libsonnet';
local util = import 'util.libsonnet';
local volumes = import 'volumes.libsonnet';
local args = import 'core-service-args.libsonnet';
local datastoreparameters = import 'datastoreparameters.libsonnet';

{
  all(metadata): {
    deployment: base.Deployment(metadata, 'core-service') {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata+: {
        namespace: metadata.namespace,
      },
      spec+: {
        template+: {
          metadata+: if metadata.enableDssMetrics then {
            annotations: {
              'prometheus.io/path': 'metrics',
              'prometheus.io/port': '8079',
              'prometheus.io/scrape': 'true',
            }
          } else {},
          spec+: {
            volumes: volumes.all(metadata).backendVolumes,
            initContainers: [
              base.WaitForDatastore(metadata),
              base.WaitForSchema(metadata, "rid"),
            ] + if metadata.enableScd then [base.WaitForSchema(metadata, "scd")] else [],
            soloContainer:: base.Container('core-service') {
              image: metadata.backend.image,
              imagePullPolicy: if metadata.backend.imagePullPolicy != '' then metadata.backend.imagePullPolicy else if metadata.cloud_provider == "minikube" then 'IfNotPresent' else 'Always',
              ports: [
                {
                  containerPort: metadata.backend.port,
                  name: 'http',
                },
              ] + if metadata.enableDssMetrics then [
                {
                  containerPort: 8079,
                  name: 'metrics',
                },
              ] else [],
              volumeMounts: volumes.all(metadata).backendMounts,
              env: util.makeEnv(metadata.backend.env),
              resources: util.resources(metadata.backend),
              command: ['core-service'],
              args_:: args.all(metadata) + datastoreparameters.all(metadata),
              readinessProbe: {
                httpGet: {
                  path: '/healthy',
                  port: metadata.backend.port,
                },
              },
            },
          },
        },
      },
    },
  },
}
