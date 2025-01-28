-- This migration drops ybgin indexes due to existing limitation to use it with other indexes.
-- https://docs.yugabyte.com/preview/explore/ysql-language-features/indexes-constraints/gin/#limitations.
DROP INDEX sc_cells_idx;
DROP INDEX ss_cells_idx;
DROP INDEX so_cells_idx;

UPDATE schema_versions set schema_version = 'v1.0.1' WHERE onerow_enforcer = TRUE;

