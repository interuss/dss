ALTER TABLE identification_service_areas DROP IF EXISTS writer;
ALTER TABLE subscriptions DROP IF EXISTS writer;
UPDATE schema_versions set schema_version = 'v3.0.0' WHERE onerow_enforcer = TRUE;