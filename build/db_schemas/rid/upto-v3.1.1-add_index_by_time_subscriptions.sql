CREATE INDEX subs_by_time_with_owner ON subscriptions (ends_at) STORING (owner);
UPDATE schema_versions set schema_version = 'v3.1.1' WHERE onerow_enforcer = TRUE;