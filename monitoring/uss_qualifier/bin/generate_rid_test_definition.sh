#!/usr/bin/env bash

set -eo pipefail

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../../.." || exit 1

monitoring/build.sh || exit 1

if [ "$CI" == "true" ]; then
  docker_args=""
else
  docker_args="-it"
fi

docker run ${docker_args} --name flight_data_generator \
  --rm \
  -e PYTHONBUFFERED=1 \
  -v "$(pwd)/monitoring/uss_qualifier/rid/test_definitions:/app/monitoring/uss_qualifier/rid/test_definitions" \
  -w /app/monitoring/uss_qualifier \
  interuss/monitoring \
  python rid/simulator/flight_state.py
