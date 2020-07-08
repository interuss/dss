DROP TABLE IF EXISTS cells_identification_service_areas;
DROP TABLE IF EXISTS cells_subscriptions;
UPDATE schema_versions set schema_version = 'v2.0.0' WHERE onerow_enforcer = TRUE