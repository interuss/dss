ALTER TABLE identification_service_areas ADD COLUMN IF NOT EXISTS cells INT64[] NOT NULL;
ALTER TABLE identification_service_areas ADD CONSTRAINT cells_not_null CHECK (array_length(cells, 1) IS NOT NULL);
CREATE INVERTED INDEX IF NOT EXISTS cell_idx on identification_service_areas (cells);
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS cells INT64[] NOT NULL;
ALTER TABLE subscriptions ADD CONSTRAINT cells_not_null CHECK (array_length(cells, 1) IS NOT NULL);
CREATE INVERTED INDEX IF NOT EXISTS cell_idx on subscriptions (cells);