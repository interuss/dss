# Database migration

Database migration is performed in a controlled fashion by defining the SQL
commands necessary to move from one database version to the next one "up" and
the commands necessary to move from that higher database version to the next one
"down".  Therefore, to make changes to the database schema, two files must be
added ("upto" and "downfrom"), prefixed also with the schema semantic version,
and named according to what the changes do.  schema_versions.schema_version
should be updated as the last step of each transition; see existing .sql files
for examples.

The two new .sql files must be added to scd.libsonnet or rid.libsonnet
(for remote ID) in this folder.

When a new database version is created, schema_versions.schema_version should be set 
in both .sql files in the appropriate folder in db_schemas. 

## Yugabyte schema versions

Versions 1.0.0 of the schemas reflect the latest versions of the crdb schemas. If 
some adaptations are required during the development phase until the first release,
changes should be done using version 1.0.1. This paragraph may be removed after the
first release.
