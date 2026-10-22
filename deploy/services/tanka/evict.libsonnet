local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';
local datastoreparameters = import 'datastoreparameters.libsonnet';
local util = import 'util.libsonnet';

{
  all(metadata): {
    local defaultMinute = util.stringToInt(metadata.locality, 60),
    SCDEvict: base.CronJob(metadata, 'dss-scd-evict') {
      spec+: {
        schedule: if metadata.evict.scd.schedule != "" then metadata.evict.scd.schedule else "%d 2 * * *" % defaultMinute,
        suspend: !metadata.evict.scd.enable_cron,
        concurrencyPolicy: "Forbid",
        jobTemplate: {
          spec: {
            template: {
              spec: {
                volumes: volumes.all(metadata).backendVolumes,
                restartPolicy: "Never",
                initContainers: [
                  base.WaitForDatastore(metadata),
                  base.WaitForSchema(metadata, "scd"),
                ],
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
                      delete: true,
                  } + (if metadata.evict.scd.timeout != "" then { timeout: metadata.evict.scd.timeout } else {})
                    + (if metadata.evict.scd.limit > 0 then { scd_limit: metadata.evict.scd.limit } else {})
                    + datastoreparameters.all(metadata),
                  volumeMounts: volumes.all(metadata).backendMounts,
                }],
              },
            },
          },
        },
      },
    },
    RIDEvict: base.CronJob(metadata, 'dss-rid-evict') {
      spec+: {
        schedule: if metadata.evict.rid.schedule != "" then metadata.evict.rid.schedule else "%d * * * *" % defaultMinute,
        suspend: !metadata.evict.rid.enable_cron,
        concurrencyPolicy: "Forbid",
        jobTemplate: {
          spec: {
            template: {
              spec: {
                volumes: volumes.all(metadata).backendVolumes,
                restartPolicy: "Never",
                initContainers: [
                  base.WaitForDatastore(metadata),
                  base.WaitForSchema(metadata, "rid"),
                ],
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
                      delete: true,
                  } + (if metadata.evict.rid.timeout != "" then { timeout: metadata.evict.rid.timeout } else {})
                    + (if metadata.evict.rid.limit > 0 then { rid_limit: metadata.evict.rid.limit } else {})
                    + datastoreparameters.all(metadata),
                  volumeMounts: volumes.all(metadata).backendMounts,
                }],
              },
            },
          },
        },
      },
    },
  },
}
