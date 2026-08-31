DROP INDEX IF EXISTS identification_service_areas@owner_idx;
DROP INDEX IF EXISTS identification_service_areas@starts_at_idx;
DROP INDEX IF EXISTS subscriptions@starts_at_idx;
UPDATE schema_versions SET schema_version = 'v4.1.0' WHERE onerow_enforcer = TRUE;
