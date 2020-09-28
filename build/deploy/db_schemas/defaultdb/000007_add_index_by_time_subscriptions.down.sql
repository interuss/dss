DROP INDEX IF EXISTS subscriptions@subs_by_time_with_owner;
UPDATE schema_versions set schema_version = 'v3.1.0' WHERE onerow_enforcer = TRUE;