DROP INDEX IF EXISTS scd_operations@owner_idx;
DROP INDEX IF EXISTS scd_operations@starts_at_idx;
DROP INDEX IF EXISTS scd_operations@altitude_lower_idx;
DROP INDEX IF EXISTS scd_operations@altitude_upper_idx;
DROP INDEX IF EXISTS scd_subscriptions@owner_idx;
DROP INDEX IF EXISTS scd_subscriptions@starts_at_idx;
DROP INDEX IF EXISTS scd_constraints@starts_at_idx;
DROP INDEX IF EXISTS scd_constraints@owner_idx;
UPDATE schema_versions SET schema_version = 'v3.4.0' WHERE onerow_enforcer = TRUE;
