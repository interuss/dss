CREATE TABLE subscriptions (
    id UUID PRIMARY KEY,
    owner TEXT NOT NULL,
    url TEXT NOT NULL,
    notification_index INT4 DEFAULT 0,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL,
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);
CREATE INDEX s_owner_idx ON subscriptions (owner);
CREATE INDEX s_starts_at_idx ON subscriptions (starts_at);
CREATE INDEX s_ends_at_idx ON subscriptions (ends_at);

CREATE TABLE cells_subscriptions (
		cell_id BIGINT NOT NULL, -- INT64 in CRDB.
		cell_level INT CHECK (cell_level BETWEEN 0 and 30),
		subscription_id UUID NOT NULL REFERENCES subscriptions (id) ON DELETE CASCADE,
		PRIMARY KEY (cell_id, subscription_id)
);
CREATE INDEX sc_cell_id_idx ON cells_subscriptions (cell_id);
CREATE INDEX sc_subscription_id_idx ON cells_subscriptions (subscription_id);

CREATE TABLE identification_service_areas (
    id UUID PRIMARY KEY,
    owner TEXT NOT NULL,
    url TEXT NOT NULL,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL,
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);
CREATE INDEX isa_owner_idx ON identification_service_areas (owner);
CREATE INDEX isa_starts_at_idx ON identification_service_areas (starts_at);
CREATE INDEX isa_ends_at_idx ON identification_service_areas (ends_at);
CREATE INDEX isa_updated_at_idx ON identification_service_areas (updated_at);

CREATE TABLE cells_identification_service_areas (
    cell_id BIGINT NOT NULL, -- INT64 in CRDB.
    cell_level INT CHECK (cell_level BETWEEN 0 and 30),
    identification_service_area_id UUID NOT NULL,
    CONSTRAINT fk_cisa_isa_id FOREIGN KEY (identification_service_area_id) REFERENCES identification_service_areas (id) ON DELETE CASCADE,
    PRIMARY KEY (cell_id, identification_service_area_id)
);
CREATE INDEX cisa_cell_id_idx ON cells_identification_service_areas (cell_id);
CREATE INDEX cisa_identification_service_area_id_idx ON cells_identification_service_areas (identification_service_area_id);

CREATE TABLE schema_versions (
	onerow_enforcer bool PRIMARY KEY DEFAULT TRUE CHECK(onerow_enforcer),
	schema_version TEXT NOT NULL
);

INSERT INTO schema_versions (schema_version) VALUES ('v1.0.0');
