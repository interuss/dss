-- This migration is equivalent to rid v4.0.0 schema for CockroachDB.

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY,
    owner TEXT NOT NULL,
    url TEXT NOT NULL,
    notification_index INT4 DEFAULT 0,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL,
    cells BIGINT[] NOT NULL,
    writer TEXT,
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at),
    CHECK (array_length(cells, 1) IS NOT NULL)
);
CREATE INDEX s_owner_idx ON subscriptions (owner);
CREATE INDEX s_starts_at_idx ON subscriptions (starts_at);
CREATE INDEX s_ends_at_idx ON subscriptions (ends_at);
CREATE INDEX s_cell_idx ON subscriptions USING ybgin (cells);
CREATE INDEX subs_by_time_with_owner ON subscriptions (ends_at) INCLUDE (owner);

CREATE TABLE identification_service_areas (
    id UUID PRIMARY KEY,
    owner TEXT NOT NULL,
    url TEXT NOT NULL,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL,
    cells BIGINT[],
    writer TEXT,
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at),
    CHECK (array_length(cells, 1) IS NOT NULL)
);
CREATE INDEX isa_owner_idx ON identification_service_areas (owner);
CREATE INDEX isa_starts_at_idx ON identification_service_areas (starts_at);
CREATE INDEX isa_ends_at_idx ON identification_service_areas (ends_at);
CREATE INDEX isa_updated_at_idx ON identification_service_areas (updated_at);
CREATE INDEX isa_cell_idx ON identification_service_areas USING ybgin (cells);

CREATE TABLE schema_versions (
	onerow_enforcer bool PRIMARY KEY DEFAULT TRUE CHECK(onerow_enforcer),
	schema_version TEXT NOT NULL
);

INSERT INTO schema_versions (schema_version) VALUES ('v1.0.0');
