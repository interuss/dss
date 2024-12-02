-- Restore cells table
CREATE TABLE IF NOT EXISTS cells_identification_service_areas (
    cell_id BIGINT NOT NULL, -- INT64 in CRDB.
    cell_level INT CHECK (cell_level BETWEEN 0 and 30),
    identification_service_area_id UUID NOT NULL,
    CONSTRAINT fk_cisa_isa_id FOREIGN KEY (identification_service_area_id) REFERENCES identification_service_areas (id) ON DELETE CASCADE,
    PRIMARY KEY (cell_id, identification_service_area_id)
);
CREATE INDEX cisa_cell_id_idx ON cells_identification_service_areas (cell_id);
CREATE INDEX cisa_identification_service_area_id_idx ON cells_identification_service_areas (identification_service_area_id);

CREATE TABLE IF NOT EXISTS cells_subscriptions (
    cell_id BIGINT NOT NULL, -- INT64 in CRDB.
    cell_level INT CHECK (cell_level BETWEEN 0 and 30),
    subscription_id UUID NOT NULL REFERENCES subscriptions (id) ON DELETE CASCADE,
    PRIMARY KEY (cell_id, subscription_id)
);
CREATE INDEX sc_cell_id_idx ON cells_subscriptions (cell_id);
CREATE INDEX sc_subscription_id_idx ON cells_subscriptions (subscription_id);

-- Migrate data from index
INSERT INTO cells_identification_service_areas
SELECT UNNEST(cells) as cell_id,
       13 AS cell_level,
       id AS identification_service_area_id
FROM identification_service_areas
ON CONFLICT (identification_service_area_id, cell_id)
DO NOTHING;

INSERT INTO cells_subscriptions
SELECT UNNEST(cells) AS cell_id,
       13 AS cell_level,
       id AS subscription_id
FROM subscriptions
ON CONFLICT (subscription_id, cell_id)
DO NOTHING;

-- Remove inverted indices
DROP INDEX IF EXISTS isa_cell_idx;
DROP INDEX IF EXISTS s_cell_idx;
ALTER TABLE identification_service_areas DROP IF EXISTS cells;
ALTER TABLE subscriptions DROP IF EXISTS cells;
UPDATE schema_versions set schema_version = 'v1.0.0' WHERE onerow_enforcer = TRUE;
