DROP INDEX IF EXISTS identification_service_areas@cell_idx;
DROP INDEX IF EXISTS subscriptions@cell_idx;
ALTER TABLE identification_service_areas DROP IF EXISTS cells;
ALTER TABLE subscriptions DROP IF EXISTS cells;
UPDATE schema_versions set schema_version = 'v1.0.0' WHERE onerow_enforcer = TRUE