CREATE TABLE IF NOT EXISTS pool_participants (
    locality TEXT PRIMARY KEY,
    public_endpoint TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS schema_versions (
  onerow_enforcer bool PRIMARY KEY DEFAULT TRUE CHECK(onerow_enforcer),
  schema_version TEXT NOT NULL
);

INSERT INTO schema_versions (schema_version) VALUES ('v1.0.0');
