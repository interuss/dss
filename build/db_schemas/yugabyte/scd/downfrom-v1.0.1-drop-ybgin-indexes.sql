CREATE INDEX sc_cells_idx ON scd_constraints USING ybgin (cells);
CREATE INDEX ss_cells_idx ON scd_subscriptions USING ybgin (cells);
CREATE INDEX so_cells_idx ON scd_operations USING ybgin (cells);
