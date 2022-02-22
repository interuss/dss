#!/usr/bin/env bash
# This script builds and executes the uss qualifier.

set -eo pipefail

if [[ $# == 0 ]]; then
  echo "Usage: $0 <CONFIG_LOCATION> [AUTH]"
  echo "Builds and executes the uss qualifier"
  echo "<CONFIG_LOCATION>: Location of the configuration file."
  echo "[AUTH]: Location of the configuration file."
  exit 1
fi
# Get absolute path of the config file provided
CONFIG_LOCATION="$(cd "$(dirname "$1")"; pwd)/$(basename "$1")"

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../../.." || exit 1

CONFIG_CONTAINER_LOCATION='/app/monitoring/uss_qualifier/config.json' # Path of the file in the container
AUTH="${2:-NoAuth()}"

QUALIFIER_OPTIONS="--auth $AUTH --config $CONFIG_CONTAINER_LOCATION"

REPORT_RID_FILE="$(pwd)/monitoring/uss_qualifier/report_rid.json"
REPORT_SCD_FILE="$(pwd)/monitoring/uss_qualifier/report_scd.json"
# files must already exist to share correctly with the Docker container
touch "${REPORT_RID_FILE}"
touch "${REPORT_SCD_FILE}"

$(pwd)/monitoring/uss_qualifier/bin/build.sh

if [ "$CI" == "true" ]; then
  docker_args="--add-host host.docker.internal:host-gateway" # Required to reach other containers in Ubuntu (used for Github Actions)
else
  docker_args="-it"
fi

docker run ${docker_args} --name uss_qualifier \
  --rm \
  -e QUALIFIER_OPTIONS="${QUALIFIER_OPTIONS}" \
  -e PYTHONBUFFERED=1 \
  -v "${REPORT_RID_FILE}:/app/monitoring/uss_qualifier/report.json" \
  -v "${REPORT_SCD_FILE}:/app/monitoring/uss_qualifier/report_scd.json" \
  -v "${CONFIG_LOCATION}:$CONFIG_CONTAINER_LOCATION" \
  interuss/uss_qualifier \
  python main.py $QUALIFIER_OPTIONS
