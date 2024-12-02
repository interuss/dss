ALTER TABLE identification_service_areas ADD COLUMN IF NOT EXISTS writer TEXT;
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS writer TEXT;
UPDATE schema_versions set schema_version = 'v3.1.0' WHERE onerow_enforcer = TRUE;
