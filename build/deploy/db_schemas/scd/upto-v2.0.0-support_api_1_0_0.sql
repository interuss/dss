/* Note: Subscription version column is now ignored; version, like OVN for
   operational intent, is encoded in updated_at */

/* Add tracking for operational intent state */
CREATE TYPE operational_intent_state AS ENUM ('Unknown', 'Accepted', 'Activated', 'Nonconforming', 'Contingent');
ALTER TABLE scd_operations ADD COLUMN state operational_intent_state NOT NULL DEFAULT 'Unknown';

/* Make Subscription associated with operational intent optional */
ALTER TABLE scd_operations ALTER COLUMN subscription_id DROP NOT NULL;

/* Record new database version */
UPDATE schema_versions set schema_version = 'v2.0.0' WHERE onerow_enforcer = TRUE;
