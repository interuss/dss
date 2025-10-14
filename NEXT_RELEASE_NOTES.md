# Upcoming Release Notes

## About & Process

This file aggregates the changes and updates that will be included in the next release notes.

Its goal is to facilitate the process of writing release notes, as well as making it easier to use this repository from its `master` branch.

Pull requests that introduce major changes, especially breaking changes to configurations, or otherwise important new features, should update this file
with the details necessary for users to migrate and fully use any added functionality.

At the time of release, the content below the horizontal line in this file will be copied to the release notes and deleted from this file.

### Template & Examples

The release notes should contain at least the following sections:

#### Mandatory migration tasks

* This version requires a database schema update. The migration is backward compatible with DSS `vX.Y.Z`.
  * For Tanka, in main.jsonnet, update desired_scd_db_version to `x.y.z`.
  * For Helm, upgrading a deployed chart with this new version will automatically migrate the version of the schema to `x.y.z`.

#### Optional migration tasks

* [terraform] cockroachdb.image.tag key with the image tag value v21.2.7 should be added to user's values.yaml files since default will be removed
  in a future release.

#### Important information

* Feature X has changed behavior to Y

#### Minimal database schema version

| Schema  | CockroachDB  | Yugabyte |
|---------|--------------|----------|
| RID     | vX.Y.Z       | vX.Y.Z   |
| SCD     | vX.Y.Z       | vX.Y.Z   |
| AUX     | vX.Y.Z       | vX.Y.Z   |

--------------------------------------------------------------------------------------------------------------------

# Release Notes for v0.21.0

## Mandatory migration tasks

* Ensure RID schema has been migrated to version 4.0.0 since the datastore doesn't fallback anymore on the default database. If that not the case, please use the previous release and upgrade to schema 4.0.0 first.
* The RID cronjob has been moved to an external command, see [#1261](https://github.com/interuss/dss/pull/1261)
    * If you want to continue to cleanup old entries regularly, run the [evict command](cmds/db-manager/cleanup/README.md) as needed:
        * Running the following command each 30 minutes will be equivalent to the previous situation
        * `db-manager evict --rid_isa=True --rid_sub=True --rid_ttl=30m --scd_oir=False --scd_sub=False`
    * Helm charts, tanka files and terraform files has been updated with defaults that run RID cleanup as before and disable SCD cleanup
        * Please review new parameters in each module specific documentation and update them as needed.
* The test certificate `build/test-certs/auth2` public and private keys have been changed, see [#1178](https://github.com/interuss/dss/pull/1178). Please update your configuration if you used that public key.
* A new `aux` store have been added, which implies a new migration job using `schema-manager`. If you ran migration jobs manually, make sure you run the migration for this new schema. Schemas are stored in the `aux_` folder.
* Upgrade terraform version to 12.2 if lower.
* The terraform variable `crdb_hostname_suffix` has been renamed to `db_hostname_suffix`, please update your configuration accordingly.
* The terraform variable `crdb_locality` has been renamed to `locality` and is now mandatory, please update your configuration accordingly.
* [`public_endpoint`](https://github.com/interuss/dss/blob/65499665ae6e6d2f4189556cf01ff671a8275ded/docs/build.md?plain=1#L460) parameter has been added as a mandatory argument. Please set it to the public endpoint of your DSS instance in order to register this instance as a participant in the pool information.
* A immutable field has been changed in terraform / tanka
  * Please remove any `*-schema-manager-*` jobs before upgrading chart / applying tanka configuration.
  * The job is only used for migrations and can be removed.

## Optional migration tasks

* Datastores parameters have been renamed to be more vendor-agnostic. The old ones have been deprecated but will continue to work for the time being.
    * If you call the dss executable directly, you will need to update parameters as below.
    * If you use helm/tanka to deploy the DSS, parameters have been renamed internally, you don't have to do anything.

| Old parameter name         | New parameter name           | Description                                                                                     |
|----------------------------|------------------------------|-------------------------------------------------------------------------------------------------|
| cockroach_application_name | datastore_application_name   | application name for tagging the connection to the database                                     |
| cockroach_db_name          | datastore_db_name            | database name to connect to                                                                     |
| cockroach_host             | datastore_host               | database host to connect to                                                                     |
| cockroach_port             | datastore_port               | database port to connect to                                                                     |
| cockroach_ssl_mode         | datastore_ssl_mode           | database sslmode                                                                                |
| cockroach_ssl_dir          | datastore_ssl_dir            | directory to ssl certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key |
| cockroach_user             | datastore_user               | database user to authenticate as                                                                |
| max_open_conns             | datastore_max_open_conns     | maximum number of open connections to the database, default is 4                                |
| max_conn_idle_secs         | datastore_max_conn_idle_secs | maximum amount of time in seconds a connection may be idle, default is 30 seconds               |
| cockroach_max_retries      | datastore_max_retries        | maximum number of attempts to retry a query in case of contention, default is 100               |

## Important information

* Yugabyte support has been added to terraform, tanka and helm files. **You're encouraged to use Yugabyte for new DSS Pool instances.**
* The RID evict task is now deleting all expired entries instead of stopping after the first one, see [#1253](https://github.com/interuss/dss/pull/1253).
    * The initial run may take longer than expected when deleting entries that may have been accumulating.
* Published images are now signed with [sigstore](https://www.sigstore.dev/), see [how to verify it](https://interuss.github.io/dss/latest/build/#verify-signature-of-prebuilt-interuss-docker-images).
* Deployment documentation has been moved to a new [website](https://interuss.github.io/dss/latest/) instead of various README in the repository tree.

## Minimal database schema version

| Schema  | CockroachDB | Yugabyte |
|---------|-------------|----------|
| AUX     |             |          |
| RID     |             |          |
| SCD     |             |          |
