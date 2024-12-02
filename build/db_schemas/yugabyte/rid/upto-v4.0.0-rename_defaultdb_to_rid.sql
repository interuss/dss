-- The following statements are expected to executed as code since database change requires
-- a reconnection to Yugabyte
-- a reconnection to Yugabyte
-- 1. ALTER DATABASE defaultdb RENAME TO rid;
-- 2. USE rid;
-- 3. Create defaultdb as scd db expects it to exist: CREATE DATABASE defaultdb;
UPDATE schema_versions set schema_version = 'v4.0.0' WHERE onerow_enforcer = TRUE;