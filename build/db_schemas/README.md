# Database migration

Database migration is performed in a controlled fashion by defining the SQL
commands necessary to move from one database version to the next one "up" and
the commands necessary to move from that higher database version to the next one
"down".  Therefore, to make changes to the database schema, two files must be
added ("upto" and "downfrom"), prefixed also with the schema semantic version,
and named according to what the changes do.  schema_versions.schema_version
should be updated as the last step of each transition; see existing .sql files
for examples.

You will need to create as well yugabyte migrations in the same way, in `yugabyte/` folder.

The two new .sql files must be added to scd.libsonnet, rid.libsonnet or aux_.libsonnet
 in this folder.

When a new database version is created, it needs to be targeted in a number of
places:
* Both .sql files in the appropriate folder in db_schemas when setting
  schema_versions.schema_version
* [DSS main.jsonnet](../../deploy/services/tanka/examples/minimum/main.jsonnet)
* [Schema manager main.jsonnet](../../deploy/services/tanka/examples/schema_manager/main.jsonnet)
* [Minikube main.jsonnet](../../deploy/services/tanka/examples/minikube/main.jsonnet)
* /pkg/{rid|scd|aux_}/store/datastore/store.go
* /deploy/infrastructure/dependencies/terraform-commons-dss/default_latest.tf
* /deploy/services/helm-charts/dss/templates/schema-manager.yaml

You can use `update_latest_version.sh` to do that automatically.

## Yugabyte schema versions

Versions 1.0.0 of the schemas reflect the latest versions of the crdb schemas. If
some adaptations are required during the development phase until the first release,
changes should be done using version 1.0.1. This paragraph may be removed after the
first release.

`aux` versions are expected to be the same between Yugabyte and CockroachDB.
