CREATE TABLE IF NOT EXISTS scd_uss_availability (
  id TEXT PRIMARY KEY,
  availability TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

/* Update database version */
UPDATE schema_versions set schema_version = 'v3.1.0' WHERE onerow_enforcer = TRUE;
