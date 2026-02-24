CREATE TABLE IF NOT EXISTS scd_locks (
  key BIGINT PRIMARY KEY
);

INSERT INTO scd_locks (key) VALUES (0);

UPDATE schema_versions set schema_version = 'v1.1.0' WHERE onerow_enforcer = TRUE;
