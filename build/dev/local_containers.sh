#!/usr/bin/env bash
# shellcheck disable=SC2034  # consumed by scripts that source this file

OAUTH_CONTAINER="dss_sandbox-local-dss-dummy-oauth-1"

if [ "${COMPOSE_PROFILES#*"with-raft-cluster"}" != "${COMPOSE_PROFILES}" ]; then
  # Talk to the load balancer so requests get spread across all 3 raft nodes.
  CORE_SERVICE_CONTAINER="dss_sandbox-local-dss-raft-lb-1"
  declare -a localhost_containers=(
    "$CORE_SERVICE_CONTAINER"
    "dss_sandbox-local-dss-core-service-1-1"
    "dss_sandbox-local-dss-core-service-2-1"
    "dss_sandbox-local-dss-core-service-3-1"
    "$OAUTH_CONTAINER"
  )
else
  CORE_SERVICE_CONTAINER="dss_sandbox-local-dss-core-service-1"
  declare -a localhost_containers=("$CORE_SERVICE_CONTAINER" "$OAUTH_CONTAINER")
fi
