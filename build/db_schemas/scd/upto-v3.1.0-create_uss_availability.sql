CREATE TABLE IF NOT EXISTS scd_uss_availability (
  id STRING PRIMARY KEY,
  availability STRING NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

/* Update database version */
UPDATE schema_versions set schema_version = 'v3.1.0' WHERE onerow_enforcer = TRUE;
