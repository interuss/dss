DROP TABLE IF EXISTS scd_locks;

UPDATE schema_versions set schema_version = 'v1.0.0' WHERE onerow_enforcer = TRUE;
