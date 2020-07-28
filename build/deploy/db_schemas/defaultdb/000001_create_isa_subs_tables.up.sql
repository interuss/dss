CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    owner STRING NOT NULL,
    url STRING NOT NULL,
    notification_index INT4 DEFAULT 0,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL,
    INDEX owner_idx (owner),
    INDEX starts_at_idx (starts_at),
    INDEX ends_at_idx (ends_at),
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);
CREATE TABLE IF NOT EXISTS cells_subscriptions (
		cell_id INT64 NOT NULL,
		cell_level INT CHECK (cell_level BETWEEN 0 and 30),
		subscription_id UUID NOT NULL REFERENCES subscriptions (id) ON DELETE CASCADE,
		PRIMARY KEY (cell_id, subscription_id),
		INDEX cell_id_idx (cell_id),
		INDEX subscription_id_idx (subscription_id)
);
CREATE TABLE IF NOT EXISTS identification_service_areas (
    id UUID PRIMARY KEY,
    owner STRING NOT NULL,
    url STRING NOT NULL,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL,
    INDEX owner_idx (owner),
    INDEX starts_at_idx (starts_at),
    INDEX ends_at_idx (ends_at),
    INDEX updated_at_idx (updated_at),
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
);
CREATE TABLE IF NOT EXISTS cells_identification_service_areas (
    cell_id INT64 NOT NULL,
    cell_level INT CHECK (cell_level BETWEEN 0 and 30),
    identification_service_area_id UUID NOT NULL REFERENCES identification_service_areas (id) ON DELETE CASCADE,
    PRIMARY KEY (cell_id, identification_service_area_id),
    INDEX cell_id_idx (cell_id),
    INDEX identification_service_area_id_idx (identification_service_area_id)
);

CREATE TABLE IF NOT EXISTS schema_versions (
	onerow_enforcer bool PRIMARY KEY DEFAULT TRUE CHECK(onerow_enforcer),
	schema_version STRING NOT NULL
);

INSERT INTO schema_versions (schema_version) VALUES ('v1.0.0');