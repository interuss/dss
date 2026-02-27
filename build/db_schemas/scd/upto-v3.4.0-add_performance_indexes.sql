-- Add composite indexes for better performance on common SCD subscription queries
-- This helps with queries that filter by time range and cells simultaneousult

-- Composite index for temporal queries with cells intersection
CREATE INDEX IF NOT EXISTS scd_subscriptions_temporal_cells_idx ON scd_subscriptions (starts_at, ends_at) WHERE cells IS NOT NULL;

-- Index for cleanup queries (expired subscriptions)
CREATE INDEX IF NOT EXISTS scd_subscriptions_cleanup_idx ON scd_subscriptions (ends_at, updated_at) WHERE ends_at IS NOT NULL OR updated_at IS NOT NULL;

-- Index to optimize notification index updates
CREATE INDEX IF NOT EXISTS scd_subscriptions_notification_idx ON scd_subscriptions (notification_index) WHERE notification_index > 0;

-- Record new database version
UPDATE schema_versions SET schema_version = 'v3.4.0' WHERE onerow_enforcer = TRUE;