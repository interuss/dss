CREATE INDEX IF NOT EXISTS owner_idx ON scd_operations (owner);
CREATE INDEX IF NOT EXISTS starts_at_idx ON scd_operations (starts_at);
CREATE INDEX IF NOT EXISTS altitude_lower_idx ON scd_operations (altitude_lower);
CREATE INDEX IF NOT EXISTS altitude_upper_idx ON scd_operations (altitude_upper);
CREATE INDEX IF NOT EXISTS owner_idx ON scd_subscriptions (owner);
CREATE INDEX IF NOT EXISTS starts_at_idx ON scd_subscriptions (starts_at);
CREATE INDEX IF NOT EXISTS owner_idx ON scd_constraints (owner);
CREATE INDEX IF NOT EXISTS starts_at_idx ON scd_constraints (starts_at);
UPDATE schema_versions SET schema_version = 'v3.3.0' WHERE onerow_enforcer = TRUE;
