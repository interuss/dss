SET sql_safe_updates = false;
DROP DATABASE IF EXISTS defaultdb;
ALTER DATABASE rid RENAME TO defaultdb;
USE defaultdb;

UPDATE schema_versions set schema_version = 'v3.1.1' WHERE onerow_enforcer = TRUE;
