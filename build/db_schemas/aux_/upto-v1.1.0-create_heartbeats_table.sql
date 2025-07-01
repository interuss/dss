CREATE TABLE IF NOT EXISTS heartbeats (
    locality TEXT NOT NULL,
    source TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    next_expected_timestamp TIMESTAMPTZ,
    reporter TEXT,
    PRIMARY KEY (locality, source)
);


UPDATE schema_versions set schema_version = 'v1.1.0' WHERE onerow_enforcer = TRUE;
