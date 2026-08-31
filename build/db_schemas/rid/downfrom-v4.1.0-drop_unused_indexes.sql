CREATE INDEX IF NOT EXISTS owner_idx ON identification_service_areas (owner);
CREATE INDEX IF NOT EXISTS starts_at_idx ON identification_service_areas (starts_at);
CREATE INDEX IF NOT EXISTS starts_at_idx ON subscriptions (starts_at);
UPDATE schema_versions SET schema_version = 'v4.0.0' WHERE onerow_enforcer = TRUE;
