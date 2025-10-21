#!/usr/bin/env bash

set -eo pipefail
set -x

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../.." || exit 1

CORE_SERVICE_CONTAINER="dss_sandbox-local-dss-core-service-1"
OAUTH_CONTAINER="dss_sandbox-local-dss-dummy-oauth-1"
declare -a localhost_containers=("$CORE_SERVICE_CONTAINER" "$OAUTH_CONTAINER")

for container_name in "${localhost_containers[@]}"; do
	if [ "$( docker container inspect -f '{{.State.Status}}' "$container_name" )" == "running" ]; then
		echo "$container_name available!"
	else
    echo '#########################################################################'
    echo '## Prerequisite to run this command is:                                ##'
    echo '## Local DSS instance + Dummy OAuth server (/build/dev/run_locally.sh) ##'
    echo '#########################################################################'
		echo "Error: $container_name not running. Execute 'build/dev/run_locally.sh up' before running build/dev/evict_locally.sh";
		exit 1;
	fi
done

# If yugabyte container is running, assume that we're running in yugabyte mode
if [ "$( docker container inspect -f '{{.State.Status}}' "dss_sandbox-local-dss-ybdb-1" )" == "running" ]; then
    echo "Activating yugabyte options"
    export DB_HOSTNAME=local-dss-ybdb
    export DB_PORT=5433
    export DB_USERNAME=yugabyte
else
    echo "Staying with cockroachdb defaults options"
fi

if ! python test/evict/test.py; then
    echo "Evict tests did not succeed."
    exit 1
else
    echo "Evict tests succeeded."
fi
