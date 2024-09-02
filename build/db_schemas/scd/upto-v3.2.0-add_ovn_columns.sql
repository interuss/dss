ALTER TABLE scd_operations
    ADD COLUMN IF NOT EXISTS uss_requested_ovn STRING
        CHECK (uss_requested_ovn != ''), -- uss_requested_ovn must be NULL if unspecified, not an empty string
    ADD COLUMN IF NOT EXISTS past_ovns         STRING[] NOT NULL
        DEFAULT ARRAY []::STRING[]
        CHECK (
            array_position(past_ovns, NULL) IS NULL AND
            array_position(past_ovns, '') IS NULL AND
            array_position(past_ovns, uss_requested_ovn) IS NULL
            ); -- past_ovns must not contain NULL elements, empty strings or current uss_requested_ovn

UPDATE schema_versions
SET schema_version = 'v3.2.0'
WHERE onerow_enforcer = TRUE;
