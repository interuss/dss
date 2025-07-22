local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';
local datastoreparameters = import 'datastoreparameters.libsonnet';

{
  all(metadata): {
    local schema_dir = if metadata.datastore == "cockroachdb" then '/db-schemas' else '/db-schemas/yugabyte',

    RIDSchemaManager: if metadata.cockroach.shouldInit || metadata.schema_manager.enable then base.Job(metadata, 'rid-schema-manager') {
      spec+: {
        template+: {
          spec+: {
            volumes: volumes.all(metadata).schemaVolumes,
            soloContainer:: base.Container('rid-schema-manager') {
              image: metadata.schema_manager.image,
              imagePullPolicy: if metadata.cloud_provider == "minikube" then 'IfNotPresent' else 'Always',
              command: ['db-manager', 'migrate'],
              args_:: {
                db_version: metadata.schema_manager.desired_rid_db_version,
                schemas_dir: schema_dir + '/rid',
              } + datastoreparameters.all(metadata),
              volumeMounts: volumes.all(metadata).schemaMounts,
            },
          },
        },
      },
    } else null,
    SCDSchemaManager: if (metadata.cockroach.shouldInit || metadata.schema_manager.enable) && metadata.enableScd then base.Job(metadata, 'scd-schema-manager') {
      spec+: {
        template+: {
          spec+: {
            volumes: volumes.all(metadata).schemaVolumes,
            soloContainer:: base.Container('scd-schema-manager') {
              image: metadata.schema_manager.image,
              imagePullPolicy: if metadata.cloud_provider == "minikube" then 'IfNotPresent' else 'Always',
              command: ['db-manager', 'migrate'],
              args_:: {
                db_version: metadata.schema_manager.desired_scd_db_version,
                schemas_dir: schema_dir + '/scd',
              } + datastoreparameters.all(metadata),
              volumeMounts: volumes.all(metadata).schemaMounts,
            },
          },
        },
      },
    } else null,
    AuxSchemaManager: if metadata.cockroach.shouldInit || metadata.schema_manager.enable then base.Job(metadata, 'aux-schema-manager') {
      spec+: {
        template+: {
          spec+: {
            volumes: volumes.all(metadata).schemaVolumes,
            soloContainer:: base.Container('aux-schema-manager') {
              image: metadata.schema_manager.image,
              imagePullPolicy: if metadata.cloud_provider == "minikube" then 'IfNotPresent' else 'Always',
              command: ['db-manager', 'migrate'],
              args_:: {
                db_version: metadata.schema_manager.desired_aux_db_version,
                schemas_dir: schema_dir + '/aux_',
              } + datastoreparameters.all(metadata),
              volumeMounts: volumes.all(metadata).schemaMounts,
            },
          },
        },
      },
    } else null,
  },
}
