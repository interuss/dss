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

## Optional migration tasks

## Important information

## Minimal database schema version

| Schema  | CockroachDB | Yugabyte |
|---------|-------------|----------|
| AUX     |             |          |
| RID     |             |          |
| SCD     |             |          |
