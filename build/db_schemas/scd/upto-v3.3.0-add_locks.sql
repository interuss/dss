CREATE TABLE IF NOT EXISTS scd_locks (
  key INT64 PRIMARY KEY
);

INSERT INTO scd_locks (key) VALUES (0);

UPDATE schema_versions
SET schema_version = 'v3.3.0'
WHERE onerow_enforcer = TRUE;
