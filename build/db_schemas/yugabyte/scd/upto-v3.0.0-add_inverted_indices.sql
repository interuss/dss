-- /* Switch to inverted indices for all Entities' cells */
ALTER TABLE scd_subscriptions ADD COLUMN IF NOT EXISTS cells BIGINT[];
CREATE INDEX ss_cells_idx ON scd_subscriptions USING ybgin (cells);
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

ALTER TABLE scd_operations ADD COLUMN IF NOT EXISTS cells BIGINT[];
CREATE INDEX so_cells_idx ON scd_operations USING ybgin (cells);
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

DROP TABLE IF EXISTS scd_cells_operations;
DROP TABLE IF EXISTS scd_cells_subscriptions;

/* Record new database version */
UPDATE schema_versions set schema_version = 'v3.0.0' WHERE onerow_enforcer = TRUE;
