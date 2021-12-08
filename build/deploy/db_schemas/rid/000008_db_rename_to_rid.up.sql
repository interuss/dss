SET sql_safe_updates = false;
ALTER DATABASE defaultdb RENAME TO rid;
USE rid;
-- Create defaultdb as scd db expects it to exist.
CREATE DATABASE defaultdb;
UPDATE schema_versions set schema_version = 'v4.0.0' WHERE onerow_enforcer = TRUE;
