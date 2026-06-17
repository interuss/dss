DELETE FROM scd_locks WHERE key > 0;

UPDATE schema_versions
SET schema_version = 'v3.4.0'
WHERE onerow_enforcer = TRUE;
