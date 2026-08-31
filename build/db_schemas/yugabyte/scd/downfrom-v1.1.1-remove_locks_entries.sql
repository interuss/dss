DELETE FROM scd_locks WHERE key BETWEEN 1 AND 65535;

UPDATE schema_versions set schema_version = 'v1.1.0' WHERE onerow_enforcer = TRUE;
