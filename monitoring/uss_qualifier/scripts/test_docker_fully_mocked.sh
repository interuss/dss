#!/usr/bin/env bash

set -eo pipefail
set -o xtrace

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../../.." || exit 1

containers=(mock_uss_ridsp mock_uss_riddp mock_uss_scdsc dss_sandbox_local-dss-core-service_1)

echo "Ensure the environment is clean"
echo "============="
build/dev/run_locally.sh down
for container_name in "${containers[@]}"; do
  docker container kill "$container_name" || echo "No pre-existing $container_name"
done

function cleanup() {
  echo "Clean up"
  echo "============="
  for container_name in "${containers[@]}"; do
    docker container kill "$container_name"
  done

  build/dev/run_locally.sh down
}

function on_exit() {
	cleanup
}

function on_sigint() {
	cleanup
	exit
}

trap on_exit   EXIT
trap on_sigint SIGINT

echo "Start mock system"
echo "============="
build/dev/run_locally.sh up -d
monitoring/mock_uss/run_locally_ridsp.sh -d
monitoring/mock_uss/run_locally_riddp.sh -d
monitoring/mock_uss/run_locally_scdsc.sh -d

echo "Wait for system to be healthy"
echo "============="
for container_name in "${containers[@]}"; do
    retry=0
    max_retry=6
    until [ "$(docker inspect -f \{\{.State.Health.Status\}\} "${container_name}")" == "healthy" ]; do
        if [ "$retry" -gt "$max_retry" ]; then
            echo "$container_name logs:"
            docker logs "$container_name"
            echo "$container_name didn't properly start. Exit." && exit 1
        fi
        echo "Waiting for $container_name to become healthy..."
        sleep 10
        retry=$((retry+1))
    done
done

echo "Run the standard local tests."
echo "============="
monitoring/uss_qualifier/run_locally.sh
echo "Run the SCD qualifier."
echo "============="
monitoring/uss_qualifier/run_locally_scd.sh
