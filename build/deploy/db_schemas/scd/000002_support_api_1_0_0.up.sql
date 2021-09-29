/* Switch Subscription version from integer to string */
SET enable_experimental_alter_column_type_general = true;
ALTER TABLE scd_subscriptions ALTER COLUMN version SET DATA TYPE STRING;

/* Add tracking for operational intent state */
CREATE TYPE operational_intent_state AS ENUM ('Unknown', 'Accepted', 'Activated', 'Nonconforming', 'Contingent');
ALTER TABLE scd_operations ADD COLUMN state operational_intent_state NOT NULL DEFAULT 'Unknown';

/* Make Subscription associated with operational intent optional */
ALTER TABLE scd_operations ALTER COLUMN subscription_id DROP NOT NULL;

/* Switch to inverted indices for all Entities' cells */
ALTER TABLE scd_subscriptions ADD COLUMN IF NOT EXISTS cells INT64[];
CREATE INVERTED INDEX IF NOT EXISTS cell_idx on scd_subscriptions (cells);
BEGIN;

WITH compact_subscription_cells AS
    ( SELECT subscription_id,
             array_agg(cell_id) AS cell_ids
     FROM scd_cells_subscriptions
     GROUP BY subscription_id)
UPDATE scd_subscriptions subscription
SET cells = compact_subscription_cells.cell_ids
FROM compact_subscription_cells
WHERE subscription.id = compact_subscription_cells.subscription_id
    AND cells IS NULL;

COMMIT;

ALTER TABLE scd_operations ADD COLUMN IF NOT EXISTS cells INT64[];
CREATE INVERTED INDEX IF NOT EXISTS cell_idx on scd_operations (cells);
BEGIN;

WITH compact_operation_cells AS
    ( SELECT operation_id,
             array_agg(cell_id) AS cell_ids
     FROM scd_cells_operations
     GROUP BY operation_id)
UPDATE scd_operations operation
SET cells = compact_operation_cells.cell_ids
FROM compact_operation_cells
WHERE operation.id = compact_operation_cells.operation_id
    AND cells IS NULL;

COMMIT;

ALTER TABLE scd_constraints ADD COLUMN IF NOT EXISTS cells INT64[];
CREATE INVERTED INDEX IF NOT EXISTS cell_idx on scd_constraints (cells);
BEGIN;

WITH compact_constraint_cells AS
    ( SELECT constraint_id,
             array_agg(cell_id) AS cell_ids
     FROM scd_cells_constraints
     GROUP BY constraint_id)
UPDATE scd_constraints constraint
SET cells = compact_constraint_cells.cell_ids
FROM compact_constraint_cells
WHERE constraint.id = compact_constraint_cells.constraint_id
    AND cells IS NULL;

COMMIT;

DROP TABLE IF EXISTS scd_cells_operations;
DROP TABLE IF EXISTS scd_cells_constraints;
DROP TABLE IF EXISTS scd_cells_subscriptions;

/* Record new database version */
UPDATE schema_versions set schema_version = 'v2.0.0' WHERE onerow_enforcer = TRUE;
