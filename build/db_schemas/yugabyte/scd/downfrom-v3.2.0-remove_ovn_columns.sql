ALTER TABLE scd_operations
    DROP IF EXISTS uss_requested_ovn,
    DROP IF EXISTS past_ovns;

UPDATE schema_versions SET schema_version = 'v3.1.0' WHERE onerow_enforcer = TRUE;
