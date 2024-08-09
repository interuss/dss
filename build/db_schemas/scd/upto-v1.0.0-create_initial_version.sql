CREATE TABLE IF NOT EXISTS scd_subscriptions (
  id UUID PRIMARY KEY,
  owner STRING NOT NULL,
  version INT4 NOT NULL DEFAULT 0,
  url STRING NOT NULL,
  notification_index INT4 DEFAULT 0,
  notify_for_operations BOOL DEFAULT false,
  notify_for_constraints BOOL DEFAULT false,
  implicit BOOL DEFAULT false,
  starts_at TIMESTAMPTZ,
  ends_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL,
  INDEX owner_idx (owner),
  INDEX starts_at_idx (starts_at),
  INDEX ends_at_idx (ends_at),
  CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at),
  CHECK (notify_for_operations OR notify_for_constraints)
);
CREATE TABLE IF NOT EXISTS scd_cells_subscriptions (
  cell_id INT64 NOT NULL,
  cell_level INT CHECK (cell_level BETWEEN 0 and 30),
  subscription_id UUID NOT NULL REFERENCES scd_subscriptions (id) ON DELETE CASCADE,
  PRIMARY KEY (cell_id, subscription_id),
  INDEX cell_id_idx (cell_id),
  INDEX subscription_id_idx (subscription_id)
);
CREATE TABLE IF NOT EXISTS scd_operations (
  id UUID PRIMARY KEY,
  owner STRING NOT NULL,
  version INT4 NOT NULL DEFAULT 0,
  url STRING NOT NULL,
  altitude_lower REAL,
  altitude_upper REAL,
  starts_at TIMESTAMPTZ,
  ends_at TIMESTAMPTZ,
  subscription_id UUID NOT NULL REFERENCES scd_subscriptions(id) ON DELETE CASCADE,
  updated_at TIMESTAMPTZ NOT NULL,
  INDEX owner_idx (owner),
  INDEX altitude_lower_idx (altitude_lower),
  INDEX altitude_upper_idx (altitude_upper),
  INDEX starts_at_idx (starts_at),
  INDEX ends_at_idx (ends_at),
  INDEX updated_at_idx (updated_at),
  INDEX subscription_id_idx (subscription_id),
  CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);
CREATE TABLE IF NOT EXISTS scd_cells_operations (
  cell_id INT64 NOT NULL,
  cell_level INT CHECK (cell_level BETWEEN 0 and 30),
  operation_id UUID NOT NULL REFERENCES scd_operations (id) ON DELETE CASCADE,
  PRIMARY KEY (cell_id, operation_id),
  INDEX cell_id_idx (cell_id),
  INDEX operation_id_idx (operation_id)
);
CREATE TABLE IF NOT EXISTS scd_constraints (
  id UUID PRIMARY KEY,
  owner STRING NOT NULL,
  version INT4 NOT NULL DEFAULT 0,
  url STRING NOT NULL,
  altitude_lower REAL,
  altitude_upper REAL,
  starts_at TIMESTAMPTZ,
  ends_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL,
  cells INT64[] NOT NULL CHECK (array_length(cells, 1) IS NOT NULL),
  INVERTED INDEX cells_idx (cells),
  INDEX owner_idx (owner),
  INDEX starts_at_idx (starts_at),
  INDEX ends_at_idx (ends_at),
  CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);

CREATE TABLE IF NOT EXISTS schema_versions (
	onerow_enforcer bool PRIMARY KEY DEFAULT TRUE CHECK(onerow_enforcer),
	schema_version STRING NOT NULL
);

INSERT INTO schema_versions (schema_version) VALUES ('v1.0.0');
