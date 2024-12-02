DROP TABLE IF EXISTS scd_uss_availability;

UPDATE schema_versions set schema_version = 'v3.0.0' WHERE onerow_enforcer = TRUE;
