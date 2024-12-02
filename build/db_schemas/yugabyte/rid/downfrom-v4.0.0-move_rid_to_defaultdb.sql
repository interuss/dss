-- The following statements are expected to executed as code since database change requires
-- a reconnection to Yugabyte
-- 1. DROP DATABASE IF EXISTS defaultdb;
-- 2. ALTER DATABASE defaultdb RENAME TO rid;
-- 3. USE defaultdb;

UPDATE schema_versions set schema_version = 'v3.1.1' WHERE onerow_enforcer = TRUE;
