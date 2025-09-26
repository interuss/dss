local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';
local datastoreparameters = import 'datastoreparameters.libsonnet';

{
  all(metadata): {
    SCDEvict: base.CronJob(metadata, 'dss-scd-evict') {
      spec+: {
        schedule: metadata.evict.scd.schedule,
        suspend: !metadata.evict.scd.enable_cron,
        concurrencyPolicy: "Forbid",
        jobTemplate: {
          spec: {
            template: {
              spec: {
                volumes: volumes.all(metadata).schemaVolumes,
                restartPolicy: "Never",
                containers: [base.Container('dss-scd-evict') {
                  image: metadata.schema_manager.image,
                  imagePullPolicy: if metadata.cloud_provider == "minikube" then 'IfNotPresent' else 'Always',
                  command: ['db-manager', 'evict'],
                  args_:: {
                      scd_oir: metadata.evict.scd.operational_intents,
                      scd_sub: metadata.evict.scd.subscriptions,
                      rid_isa: false,
                      rid_sub: false,
                      scd_ttl: metadata.evict.scd.ttl,
                  } + datastoreparameters.all(metadata),
                  volumeMounts: volumes.all(metadata).schemaMounts,
                }],
              },
            },
          },
        },
      },
    },
    RIDEvict: base.CronJob(metadata, 'dss-rid-evict') {
      spec+: {
        schedule: metadata.evict.rid.schedule,
        suspend: !metadata.evict.rid.enable_cron,
        concurrencyPolicy: "Forbid",
        jobTemplate: {
          spec: {
            template: {
              spec: {
                volumes: volumes.all(metadata).schemaVolumes,
                restartPolicy: "Never",
                containers: [base.Container('dss-rid-evict') {
                  image: metadata.schema_manager.image,
                  imagePullPolicy: if metadata.cloud_provider == "minikube" then 'IfNotPresent' else 'Always',
                  command: ['db-manager', 'evict'],
                  args_:: {
                      scd_oir: false,
                      scd_sub: false,
                      rid_isa: metadata.evict.rid.ISAs,
                      rid_sub: metadata.evict.rid.subscriptions,
                      rid_ttl: metadata.evict.rid.ttl,
                  } + datastoreparameters.all(metadata),
                  volumeMounts: volumes.all(metadata).schemaMounts,
                }],
              },
            },
          },
        },
      },
    },
  },
}
