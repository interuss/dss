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