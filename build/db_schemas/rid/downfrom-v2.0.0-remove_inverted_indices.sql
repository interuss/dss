-- Restore cells table
CREATE TABLE IF NOT EXISTS cells_identification_service_areas (
    cell_id INT64 NOT NULL,
    cell_level INT CHECK (cell_level BETWEEN 0 and 30),
    identification_service_area_id UUID NOT NULL REFERENCES identification_service_areas (id) ON DELETE CASCADE,
    PRIMARY KEY (cell_id, identification_service_area_id),
    INDEX cell_id_idx (cell_id),
    INDEX identification_service_area_id_idx (identification_service_area_id)
);

CREATE TABLE IF NOT EXISTS cells_subscriptions (
		cell_id INT64 NOT NULL,
		cell_level INT CHECK (cell_level BETWEEN 0 and 30),
		subscription_id UUID NOT NULL REFERENCES subscriptions (id) ON DELETE CASCADE,
		PRIMARY KEY (cell_id, subscription_id),
		INDEX cell_id_idx (cell_id),
		INDEX subscription_id_idx (subscription_id)
);

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
DROP INDEX IF EXISTS identification_service_areas@cell_idx;
DROP INDEX IF EXISTS subscriptions@cell_idx;
ALTER TABLE identification_service_areas DROP IF EXISTS cells;
ALTER TABLE subscriptions DROP IF EXISTS cells;
UPDATE schema_versions set schema_version = 'v1.0.0' WHERE onerow_enforcer = TRUE;
