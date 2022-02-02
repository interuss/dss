-- Restore Table scd_cells_operations
CREATE TABLE IF NOT EXISTS scd_cells_operations (
  cell_id INT64 NOT NULL,
  cell_level INT CHECK (cell_level BETWEEN 0 and 30),
  operation_id UUID NOT NULL REFERENCES scd_operations (id) ON DELETE CASCADE,
  PRIMARY KEY (cell_id, operation_id),
  INDEX cell_id_idx (cell_id),
  INDEX operation_id_idx (operation_id)
);

-- Restore Table scd_cells_subscriptions
CREATE TABLE IF NOT EXISTS scd_cells_subscriptions (
  cell_id INT64 NOT NULL,
  cell_level INT CHECK (cell_level BETWEEN 0 and 30),
  subscription_id UUID NOT NULL REFERENCES scd_subscriptions (id) ON DELETE CASCADE,
  PRIMARY KEY (cell_id, subscription_id),
  INDEX cell_id_idx (cell_id),
  INDEX subscription_id_idx (subscription_id)
);
-- Remove inverted index
DROP INDEX IF EXISTS scd_subscriptions@cell_idx;
ALTER TABLE scd_operations DROP IF EXISTS cells;

UPDATE schema_versions set schema_version = 'v2.0.0' WHERE onerow_enforcer = TRUE;
