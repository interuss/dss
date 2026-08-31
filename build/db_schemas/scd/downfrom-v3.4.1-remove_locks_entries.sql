DELETE FROM scd_locks WHERE key BETWEEN 1 AND 65535;

UPDATE schema_versions
SET schema_version = 'v3.4.0'
WHERE onerow_enforcer = TRUE;
