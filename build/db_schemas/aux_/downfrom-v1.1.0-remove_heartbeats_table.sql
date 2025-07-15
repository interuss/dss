DROP TABLE IF EXISTS heartbeats;
UPDATE schema_versions set schema_version = 'v1.0.0' WHERE onerow_enforcer = TRUE;
