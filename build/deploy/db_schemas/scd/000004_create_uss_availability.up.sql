CREATE TABLE IF NOT EXISTS uss_availability (
  id STRING PRIMARY KEY,
  availability STRING NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
)