local base = import 'base.libsonnet';
local volumes = import 'volumes.libsonnet';

local schema_dir = '/db-schemas';

{
  all(metadata): {
    assert metadata.cockroach.shouldInit == true && metadata.cockroach.JoinExisting == [] : "If shouldInit is True, JoinExisiting should be empty",
    RIDSchemaManager: if metadata.cockroach.shouldInit then base.Job(metadata, 'rid-schema-manager') {
      spec+: {
        template+: {
          spec+: {
            volumes_: {
              client_certs: volumes.volumes.client_certs,
              ca_certs: volumes.volumes.ca_certs,
            },
            soloContainer:: base.Container('rid-schema-manager') {
              image: metadata.schema_manager.image,
              command: ['db-manager'],
              args_:: {
                cockroach_host: 'cockroachdb-balanced.' + metadata.namespace,
                cockroach_port: metadata.cockroach.grpc_port,
                cockroach_ssl_mode: 'verify-full',
                cockroach_user: 'root',
                cockroach_ssl_dir: '/cockroach/cockroach-certs',
                db_version: metadata.schema_manager.desired_rid_db_version,
                schemas_dir: schema_dir + '/rid',
              },
              volumeMounts: volumes.mounts.caCert + volumes.mounts.clientCert,
            },
            imagePullSecrets: if metadata.image_pull_secret != null then [{name: metadata.image_pull_secret}] else [],
          },
        },
      },
    } else null,
    SCDSchemaManager: if metadata.cockroach.shouldInit && metadata.enableScd then base.Job(metadata, 'scd-schema-manager') {
      spec+: {
        template+: {
          spec+: {
            volumes_: {
              client_certs: volumes.volumes.client_certs,
              ca_certs: volumes.volumes.ca_certs,
            },
            soloContainer:: base.Container('scd-schema-manager') {
              image: metadata.schema_manager.image,
              command: ['db-manager'],
              args_:: {
                cockroach_host: 'cockroachdb-balanced.' + metadata.namespace,
                cockroach_port: metadata.cockroach.grpc_port,
                cockroach_ssl_mode: 'verify-full',
                cockroach_user: 'root',
                cockroach_ssl_dir: '/cockroach/cockroach-certs',
                db_version: metadata.schema_manager.desired_scd_db_version,
                schemas_dir: schema_dir + '/scd',
              },
              volumeMounts: volumes.mounts.caCert + volumes.mounts.clientCert,
            },
            imagePullSecrets: if metadata.image_pull_secret != null then [{name: metadata.image_pull_secret}] else [],
          },
        },
      },
    } else null,
  },
}
