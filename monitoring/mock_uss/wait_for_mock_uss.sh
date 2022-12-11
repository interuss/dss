#!/bin/bash

MOCK_USS_CONTAINER="${1:?The mock_uss container name must be specified (e.g., wait_for_mock_uss.sh mock_uss_scdsc)}"

n_delays=0
max_delays=20

# Check that container is running
last_message=""
while true; do
  if [ "$( docker container inspect -f '{{.State.Status}}' "${MOCK_USS_CONTAINER}" 2>/dev/null)" = "running" ]; then
    break
  fi
  new_message="Waiting for ${MOCK_USS_CONTAINER} container to start..."
  if [ "${new_message}" = "${last_message}" ]; then
    printf "."
  else
    printf '%s' "${new_message}"
    last_message="${new_message}"
  fi
  sleep 3
  ((n_delays=n_delays+1))
  if [ $n_delays -gt $max_delays ]; then
    echo ""
    echo "Mock USS container ${MOCK_USS_CONTAINER} did not start in a reasonable amount of time"
    exit 1
  fi
done
if [ -n "${last_message}" ]; then
  echo ""
fi

last_message=""
while true; do
  health_status="$( docker container inspect -f '{{.State.Health.Status}}' "${MOCK_USS_CONTAINER}" )"
    if [ "${health_status}" = "healthy" ]; then
      break
    else
      new_message="Waiting for ${MOCK_USS_CONTAINER} to be available (currently ${health_status})..."
      if [ "${new_message}" = "${last_message}" ]; then
        printf "."
      else
        printf '%s' "${new_message}"
        last_message="${new_message}"
      fi
      sleep 3
      ((n_delays=n_delays+1))
      if [ $n_delays -gt $max_delays ]; then
        echo ""
        echo "Mock USS container ${MOCK_USS_CONTAINER} did not become healthy in a reasonable amount of time"
        exit 1
      fi
    fi
done
if [ -n "${last_message}" ]; then
  echo ""
fi

echo "Mock USS ${MOCK_USS_CONTAINER} is now available."
