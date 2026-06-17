INSERT INTO scd_locks (key)
SELECT generate_series(1, 65535)
ON CONFLICT (key) DO NOTHING;

UPDATE schema_versions
SET schema_version = 'v3.4.1'
WHERE onerow_enforcer = TRUE;
