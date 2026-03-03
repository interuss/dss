# DSS SCD Performance Optimization Guide

## Overview

This document provides guidelines for monitoring and optimizing SCD (Strategic Conflict Detection) performance in the DSS, particularly addressing high latency issues with subscription operations.

## Performance Issues and Solutions

### Issue: High Latency for SCD Subscription Operations

**Symptoms:**
- API calls exceeding 10+ seconds
- High transaction latency (99th percentile spikes)
- Database lock contention
- Transaction restarts

**Root Causes:**
1. Large `scd_subscriptions` table without proper cleanup
2. Inefficient row-level locking strategy
3. Missing or suboptimal database indexes
4. Infrequent cleanup of expired subscriptions

### Implemented Solutions

#### 1. Global Locking (Immediate Impact)

**What changed:** Enabled global locking by default (`enable_scd_global_lock=true`)

**Benefits:**
- Reduces lock contention when multiple operations access overlapping subscription areas
- Trades some global throughput for better performance under high subscription density
- Eliminates expensive `SELECT id FROM scd_subscriptions WHERE cells && $1` queries

**Monitoring:**
```bash
# Check if global locking is enabled
grep "enable_scd_global_lock" /proc/$(pgrep core-service)/cmdline

# Monitor lock wait times in logs
grep "SCD subscription locking took longer" /var/log/dss.log
```

#### 2. Performance Monitoring and Logging

**What changed:** Added performance monitoring to critical functions:
- `LockSubscriptionsOnCells`: Logs when locking takes >100ms
- `SearchSubscriptions`: Logs when searches take >50ms

**Monitoring:**
```bash
# Monitor slow operations
tail -f /var/log/dss.log | grep "took longer than expected"

# Check performance patterns
grep "SCD subscription" /var/log/dss.log | awk '{print $4}' | sort | uniq -c
```

#### 3. Database Index Optimization

**What changed:** Added composite indexes for common query patterns:
- `scd_subscriptions_temporal_cells_idx`: Temporal queries with cells
- `scd_subscriptions_cleanup_idx`: Expired subscription cleanup
- `scd_subscriptions_notification_idx`: Notification updates

**Apply the migration:**
```bash
psql -h $DB_HOST -d $DB_NAME -f build/db_schemas/scd/upto-v3.4.0-add_performance_indexes.sql
```

**Verify indexes:**
```sql
-- Check index usage
SELECT schemaname, tablename, indexname, idx_tup_read, idx_tup_fetch 
FROM pg_stat_user_indexes 
WHERE tablename = 'scd_subscriptions';

-- Check index sizes
SELECT indexname, pg_size_pretty(pg_relation_size(indexname::regclass)) 
FROM pg_indexes 
WHERE tablename = 'scd_subscriptions';
```

#### 4. Automated Cleanup

**What changed:** Created automated cleanup scripts and cron configurations

**Setup:**
```bash
# Make cleanup script executable
chmod +x scripts/cleanup_subscriptions.sh

# Test cleanup (dry run)
./scripts/cleanup_subscriptions.sh

# Setup cron (example for daily cleanup)
echo "0 2 * * * DRY_RUN=false /path/to/dss/scripts/cleanup_subscriptions.sh" | crontab -
```

## Monitoring and Maintenance

### Key Performance Metrics

1. **Subscription Table Size:**
```sql
SELECT 
    COUNT(*) as total_subscriptions,
    COUNT(*) FILTER (WHERE ends_at < NOW()) as expired_subscriptions,
    COUNT(*) FILTER (WHERE ends_at IS NULL AND updated_at < NOW() - INTERVAL '112 days') as stale_subscriptions
FROM scd_subscriptions;
```

2. **Lock Performance:**
```bash
# Monitor lock wait times
grep "SCD subscription locking" /var/log/dss.log | tail -20
```

3. **Query Performance:**
```sql
-- Check slow queries (if pg_stat_statements is enabled)
SELECT query, mean_time, calls, total_time 
FROM pg_stat_statements 
WHERE query LIKE '%scd_subscriptions%' 
ORDER BY mean_time DESC;
```

### Recommended Monitoring Thresholds

- **Subscription Count:** Alert if >5000 active subscriptions
- **Lock Duration:** Alert if >500ms consistently  
- **Search Duration:** Alert if >200ms consistently
- **Expired Subscriptions:** Alert if >10% of total subscriptions

### Cleanup Schedule Recommendations

| Environment | Table Size | Cleanup Frequency | TTL |
|-------------|------------|-------------------|-----|
| Development | <1000 | Every 2 hours | 24h |
| Staging | <3000 | Every 6 hours | 168h (7 days) |
| Production | <5000 | Daily | 2688h (112 days) |
| High Traffic | <2000 | Every 4 hours | 720h (30 days) |

### Troubleshooting

#### High Latency Persists

1. **Check Global Lock Status:**
```bash
ps aux | grep core-service | grep enable_scd_global_lock
```

2. **Monitor Database Locks:**
```sql
SELECT pid, state, query_start, state_change, query 
FROM pg_stat_activity 
WHERE query LIKE '%scd_subscriptions%' AND state != 'idle';

SELECT locktype, database, relation, mode, granted 
FROM pg_locks l 
JOIN pg_class c ON l.relation = c.oid 
WHERE c.relname = 'scd_subscriptions';
```

3. **Check Cleanup Effectiveness:**
```bash
# Review cleanup logs
tail -100 /var/log/dss-cleanup.log

# Check for cleanup failures
grep ERROR /var/log/dss-cleanup.log
```

#### Large Number of Subscriptions

If subscription count remains high after cleanup:

1. **Reduce TTL temporarily:**
```bash
DRY_RUN=false SCD_TTL=168h ./scripts/cleanup_subscriptions.sh
```

2. **Implement batched cleanup:**
```bash
# Clean in smaller batches
for ttl in 720h 1440h 2160h; do
    DRY_RUN=false SCD_TTL=$ttl ./scripts/cleanup_subscriptions.sh
    sleep 300  # Wait 5 minutes between batches
done
```

3. **Consider schema optimization:**
   - Partition large tables by time ranges
   - Archive old subscriptions to separate tables

## Performance Testing

### Load Testing Scenarios

1. **Concurrent Subscription Creation:**
```bash
# Test with multiple concurrent subscription creates
for i in {1..10}; do
    curl -X POST "https://dss/v1/dss/subscriptions" -d @test_subscription.json &
done
wait
```

2. **Search Performance:**
```bash
# Test subscription searches under load
ab -n 100 -c 10 "https://dss/v1/dss/subscriptions?area=..."
```

### Benchmarking

Before and after performance comparison:

```sql
-- Measure query performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT id FROM scd_subscriptions 
WHERE cells && ARRAY[123456789] 
AND starts_at <= NOW() 
AND ends_at >= NOW();
```

## Best Practices

1. **Regular Monitoring:** Set up automated monitoring of key metrics
2. **Proactive Cleanup:** Don't wait for performance issues to run cleanup
3. **Global Lock Usage:** Keep global lock enabled in production environments
4. **Index Maintenance:** Regular REINDEX on high-traffic systems
5. **Capacity Planning:** Monitor subscription growth trends

## Emergency Procedures

### Immediate Response to High Latency

1. **Enable Global Lock (if disabled):**
```bash
# Restart service with global lock enabled
sudo systemctl stop dss-core-service
sudo systemctl start dss-core-service --enable_scd_global_lock=true
```

2. **Emergency Cleanup:**
```bash
# Aggressive cleanup with short TTL
DRY_RUN=false SCD_TTL=24h ./scripts/cleanup_subscriptions.sh
```

3. **Database Optimization:**
```sql
-- Force statistics update
ANALYZE scd_subscriptions;

-- Rebuild indexes if needed
REINDEX TABLE scd_subscriptions;
```