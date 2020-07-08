ALTER TABLE identification_service_areas ADD COLUMN IF NOT EXISTS cells INT64[];
CREATE INVERTED INDEX IF NOT EXISTS cell_idx on identification_service_areas (cells);
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS cells INT64[];
CREATE INVERTED INDEX IF NOT EXISTS cell_idx on subscriptions (cells);