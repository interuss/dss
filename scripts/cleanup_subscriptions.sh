#!/bin/bash

# DSS SCD Subscription Cleanup Script
# This script helps automate the cleanup of expired SCD subscriptions to prevent
# the high latency issues caused by large subscription tables.

set -euo pipefail

# Configuration
SCD_TTL=${SCD_TTL:-"2688h"}  # Default: 112 days (2*56 days)
DRY_RUN=${DRY_RUN:-"true"}    # Set to "false" to actually delete
DB_HOST=${DB_HOST:-"localhost"}
DB_NAME=${DB_NAME:-"dss"}
LOG_FILE=${LOG_FILE:-"/var/log/dss-cleanup.log"}

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

# Performance monitoring
start_time=$(date +%s)

log "Starting DSS SCD subscription cleanup"
log "Configuration: SCD_TTL=$SCD_TTL, DRY_RUN=$DRY_RUN"

# Change to the DSS directory
cd "$(dirname "$0")/../.."

# Build the db-manager if it doesn't exist
if [[ ! -f "./db-manager" ]]; then
    log "Building db-manager..."
    go build -o db-manager ./cmds/db-manager
fi

# Prepare cleanup command
cleanup_cmd="./db-manager evict --scd_ttl=$SCD_TTL --datastore_host=$DB_HOST --datastore_db_name=$DB_NAME"

if [[ "$DRY_RUN" == "false" ]]; then
    cleanup_cmd="$cleanup_cmd --delete"
    log "WARNING: Running with --delete enabled. Expired subscriptions will be permanently removed."
else
    log "Running in DRY RUN mode. Use DRY_RUN=false to actually delete expired subscriptions."
fi

# Run the cleanup
log "Executing: $cleanup_cmd"
if $cleanup_cmd 2>&1 | tee -a "$LOG_FILE"; then
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    log "Cleanup completed successfully in ${duration}s"
else
    log "ERROR: Cleanup failed with exit code $?"
    exit 1
fi

log "Cleanup process finished"