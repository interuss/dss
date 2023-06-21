#!/bin/bash

OAUTH_CONTAINER="dss_sandbox-local-dss-dummy-oauth-1"
CORE_SERVICE_CONTAINER="dss_sandbox-local-dss-core-service-1"
declare -a localhost_containers=("$OAUTH_CONTAINER" "$CORE_SERVICE_CONTAINER")

# 2 minute timer to prevent infinite looping if a docker issue is present.
timeout_duration=120

# check to see if 2 minutes has elapsed which indicates a problem with the container(s)
check_timeout() {
   local start_time="$1"
   local error_message="$2"
   current_time=$(date +%s)
   elapsed_time=$((current_time-start_time))
   if ((elapsed_time >= timeout_duration)); then
     echo "$error_message"
     exit 1
   fi
}

# start the timer
start_time=$(date +%s)
for container_name in "${localhost_containers[@]}"; do
  last_message=""
  while true; do
    if [ "$( docker container inspect -f '{{.State.Status}}' "${container_name}" 2>/dev/null)" = "running" ]; then
      break
    fi
    new_message="Waiting for ${container_name} container to start..."
    if [ "${new_message}" = "${last_message}" ]; then
      printf "."
    else
      printf '%s' "${new_message}"
      last_message="${new_message}"
    fi
    check_timeout "$start_time" "Timeout reached. Container failed to start. Exiting."
    sleep 3
  done
  if [ -n "${last_message}" ]; then
    echo ""
  fi
done

# reset the timer
start_time=$(date +%s)
last_message=""
while true; do
  health_status="$( docker container inspect -f '{{.State.Health.Status}}' "${CORE_SERVICE_CONTAINER}" )"
    if [ "${health_status}" = "healthy" ]; then
      break
    else
      new_message="Waiting for ${CORE_SERVICE_CONTAINER} to be available (currently ${health_status})..."
      if [ "${new_message}" = "${last_message}" ]; then
        printf "."
      else
        printf '%s' "${new_message}"
        last_message="${new_message}"
      fi
      check_timeout "$start_time" "Timeout reached. Container failed to become available. Exiting."
      sleep 3
    fi
done
if [ -n "${last_message}" ]; then
  echo ""
fi

echo "Local DSS instance is now available."
