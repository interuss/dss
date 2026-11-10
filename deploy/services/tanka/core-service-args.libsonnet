// Common core-service CLI parameters of the SQL and Raft deployments
{
  all(metadata): {
    addr: ':' + std.toString(metadata.backend.port),
    gcp_prof_service_name: metadata.backend.prof_grpc_name,
    dump_requests: metadata.backend.dumpRequests,
    accepted_jwt_audiences: metadata.backend.hostname,
    locality: metadata.locality,
    enable_rid: metadata.enableRid,
    enable_scd: metadata.enableScd,
    enable_scd_global_lock: metadata.enableScdGlobalLock,
  }
  + (if std.join(",", metadata.backend.pubKeys) != '' then {
    public_key_files: std.join(",", metadata.backend.pubKeys),
  } else {}) + (if metadata.backend.jwksEndpoint != '' then {
    jwks_endpoint: metadata.backend.jwksEndpoint,
  } else {}) + (if metadata.backend.jwksKeyIds != [] && metadata.backend.jwksKeyIds != [''] then {
    jwks_key_ids: std.join(",", metadata.backend.jwksKeyIds),
  } else {}) + (if metadata.backend.jwksRefreshInterval != '' then {
    jwks_refresh_interval: metadata.backend.jwksRefreshInterval,
  } else {}) + (if metadata.backend.jwksKeyTtl != '' then {
    jwks_key_ttl: metadata.backend.jwksKeyTtl,
  } else {}) + (if metadata.backend.publicEndpoint != '' then {
    public_endpoint: metadata.backend.publicEndpoint,
  } else {}) + (if metadata.enableDssMetrics then {
    enable_metrics: true,
  } else {}) + (if metadata.enableTimeBasedNotificationIndex then {
    enable_time_based_notification_index: true,
  } else {})
}
