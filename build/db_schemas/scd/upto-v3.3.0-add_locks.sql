CREATE TABLE IF NOT EXISTS scd_locks (
  id INT64 PRIMARY KEY
);

INSERT INTO scd_locks (id) VALUES (0);

UPDATE schema_versions
SET schema_version = 'v3.3.0'
WHERE onerow_enforcer = TRUE;
