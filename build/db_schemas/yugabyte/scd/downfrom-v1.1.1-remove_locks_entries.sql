DELETE FROM scd_locks WHERE key > 0;

UPDATE schema_versions set schema_version = 'v1.1.0' WHERE onerow_enforcer = TRUE;
