CREATE INDEX s_cell_idx ON subscriptions USING ybgin (cells);
CREATE INDEX isa_cell_idx ON identification_service_areas USING ybgin (cells);

UPDATE schema_versions set schema_version = 'v1.0.0' WHERE onerow_enforcer = TRUE;
