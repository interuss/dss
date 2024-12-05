-- This migration is equivalent to scd v3.2.0 schema for CockroachDB.

CREATE TABLE IF NOT EXISTS scd_subscriptions (
  id UUID PRIMARY KEY,
  owner TEXT NOT NULL,
  version INT4 NOT NULL DEFAULT 0,
  url TEXT NOT NULL,
  notification_index INT4 DEFAULT 0,
  notify_for_operations BOOL DEFAULT false,
  notify_for_constraints BOOL DEFAULT false,
  implicit BOOL DEFAULT false,
  starts_at TIMESTAMPTZ,
  ends_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL,
  cells BIGINT[],
  CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at),
  CHECK (notify_for_operations OR notify_for_constraints)
);
CREATE INDEX ss_owner_idx ON scd_subscriptions (owner);
CREATE INDEX ss_starts_at_idx ON scd_subscriptions (starts_at);
CREATE INDEX ss_ends_at_idx ON scd_subscriptions (ends_at);
CREATE INDEX ss_cells_idx ON scd_subscriptions USING ybgin (cells);

CREATE TYPE operational_intent_state AS ENUM ('Unknown', 'Accepted', 'Activated', 'Nonconforming', 'Contingent');

CREATE TABLE IF NOT EXISTS scd_operations (
  id UUID PRIMARY KEY,
  owner TEXT NOT NULL,
  version INT4 NOT NULL DEFAULT 0,
  url TEXT NOT NULL,
  altitude_lower REAL,
  altitude_upper REAL,
  starts_at TIMESTAMPTZ,
  ends_at TIMESTAMPTZ,
  subscription_id UUID REFERENCES scd_subscriptions(id) ON DELETE CASCADE,
  updated_at TIMESTAMPTZ NOT NULL,
  state operational_intent_state NOT NULL DEFAULT 'Unknown',
  cells BIGINT[],
  uss_requested_ovn TEXT,
  past_ovns TEXT[] NOT NULL DEFAULT ARRAY []::TEXT[],
  CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at),
  CHECK (uss_requested_ovn != ''),
  CHECK (
      array_position(past_ovns, NULL) IS NULL AND
      array_position(past_ovns, '') IS NULL AND
      array_position(past_ovns, uss_requested_ovn) IS NULL
      )
);
CREATE INDEX so_owner_idx ON scd_operations (owner);
CREATE INDEX so_altitude_lower_idx ON scd_operations (altitude_lower);
CREATE INDEX so_altitude_upper_idx ON scd_operations (altitude_upper);
CREATE INDEX so_starts_at_idx ON scd_operations (starts_at);
CREATE INDEX so_ends_at_idx ON scd_operations (ends_at);
CREATE INDEX so_updated_at_idx ON scd_operations (updated_at);
CREATE INDEX so_subscription_id_idx ON scd_operations (subscription_id);
CREATE INDEX so_cells_idx ON scd_operations USING ybgin (cells);

CREATE TABLE IF NOT EXISTS scd_constraints (
  id UUID PRIMARY KEY,
  owner TEXT NOT NULL,
  version INT4 NOT NULL DEFAULT 0,
  url TEXT NOT NULL,
  altitude_lower REAL,
  altitude_upper REAL,
  starts_at TIMESTAMPTZ,
  ends_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ NOT NULL,
  cells BIGINT[] NOT NULL CHECK (array_length(cells, 1) IS NOT NULL),
  CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);
CREATE INDEX sc_owner_idx ON scd_constraints (owner);
CREATE INDEX sc_starts_at_idx ON scd_constraints (starts_at);
CREATE INDEX sc_ends_at_idx ON scd_constraints (ends_at);
CREATE INDEX sc_cells_idx ON scd_constraints USING ybgin (cells);


CREATE TABLE IF NOT EXISTS scd_uss_availability (
  id TEXT PRIMARY KEY,
  availability TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS schema_versions (
  onerow_enforcer bool PRIMARY KEY DEFAULT TRUE CHECK(onerow_enforcer),
  schema_version TEXT NOT NULL
);

INSERT INTO schema_versions (schema_version) VALUES ('v1.0.0');
