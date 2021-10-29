#!/usr/bin/env bash

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../.." || exit 1

docker build \
    -f monitoring/rid_qualifier/Dockerfile \
    -t interuss/dss/rid_qualifier \
    --build-arg version="$(scripts/git/commit.sh)" \
    monitoring

docker run -i -t --name flight_data_generator \
  --rm \
  --tty \
  -e PYTHONBUFFERED=1 \
  -v "$(pwd)/monitoring/rid_qualifier/test_definitions:/app/monitoring/rid_qualifier/test_definitions" \
  interuss/dss/rid_qualifier \
  python flight_data_generator.py
