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

* The RID cronjob has been moved to an external command, see [#1261](https://github.com/interuss/dss/pull/1261)
    * If you want to continue to cleanup old entries regularly, run the [evict command](cmds/db-manager/cleanup/README.md) as needed:
        * Running the following command each 30 minutes will be equivalent to the previous situation
        * `db-manager evict --rid_isa=True --rid_sub=True --rid_ttl=30m --scd_oir=False --scd_sub=False`
    * Helm charts, tanka files and terraform files has been updated with defaults that run RID cleanup as before and disable SCD cleanup
        * Please review new parameters in each module specific documentation and update them as needed.
* The `auth2` key has been changed, see [#1178](https://github.com/interuss/dss/pull/1178). Please update your configuration if you used that public key.
* A new `aux` store have been added. If you ran migration job manually, please take the new schema into account. Schemas are stored in the `aux_` folder.
* The terraform variable `crdb_hostname_suffix` has been renamed to `db_hostname_suffix`, please update your configuration accordinly.
* The terraform variable `crdb_locality` has been renamed to `locality` and is now mandatory, please update your configuration accordinly.

## Optional migration tasks

* A new parameter have been added, `public_endpoint`. Please set it to the public endpoint of your DSS instance.

## Important information

* The RID evict task is now deleting all expired entries instead of stopping after the first one, see [#1253](https://github.com/interuss/dss/pull/1253).
    * The initial run may take longer than expected when deleting entries that may have been accumulating.
* Published images are now signed with [sigstore](https://www.sigstore.dev/), see [how to verify it](https://interuss.github.io/dss/latest/build/#verify-signature-of-prebuilt-interuss-docker-images).
* The rid datastore don't fallback anymore on the default database, it's expected that you upgraded to schema 4.0.0, released 1 year ago. If that not the case, please use the previous release and upgrade to schmea 4.0.0 first.
* Yugabyte support have been added to terraform, tanka and helm files. You're encouraged to use Yugabyte for new DSS Pool instances.
* Documentation have been moved to a new [website](https://interuss.github.io/dss/latest/) instead of various README in the repository tree.

## Minimal database schema version

| Schema  | CockroachDB | Yugabyte |
|---------|-------------|----------|
| AUX     |             |          |
| RID     |             |          |
| SCD     |             |          |
