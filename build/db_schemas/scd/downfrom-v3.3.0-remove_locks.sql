DROP TABLE IF EXISTS scd_locks;

UPDATE schema_versions
SET schema_version = 'v3.2.0'
WHERE onerow_enforcer = TRUE;
