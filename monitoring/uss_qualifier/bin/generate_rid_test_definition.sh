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

docker build \
    -f monitoring/uss_qualifier/Dockerfile \
    -t interuss/uss_qualifier \
    --build-arg version="$(scripts/git/commit.sh)" \
    monitoring

if [ "$CI" == "true" ]; then
  docker_args=""
else
  docker_args="-it"
fi

docker run ${docker_args} --name flight_data_generator \
  --rm \
  -e PYTHONBUFFERED=1 \
  -v "$(pwd)/monitoring/uss_qualifier/rid/test_definitions:/app/monitoring/uss_qualifier/rid/test_definitions" \
  interuss/uss_qualifier \
  python rid/simulator/flight_state.py
