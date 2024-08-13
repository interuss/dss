ALTER TABLE IF EXISTS scd_operations DROP COLUMN state;
DROP TYPE IF EXISTS operational_intent_state;
ALTER TABLE scd_operations ALTER COLUMN subscription_id SET NOT NULL;

UPDATE schema_versions set schema_version = 'v1.0.0' WHERE onerow_enforcer = TRUE;
