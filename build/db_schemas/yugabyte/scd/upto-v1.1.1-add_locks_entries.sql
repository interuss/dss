INSERT INTO scd_locks (key)
SELECT generate_series(1, 65535)
ON CONFLICT (key) DO NOTHING;

UPDATE schema_versions set schema_version = 'v1.1.1' WHERE onerow_enforcer = TRUE;
