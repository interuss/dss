#!/usr/bin/env bash

# Find and change to repo root directory
OS=$(uname)
if [[ "$OS" == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")"
else
	BASEDIR=$(readlink -e "$(dirname "$0")")
fi
cd "${BASEDIR}/../../.." || exit 1

IN_FILE=$1   # Input filename.

if ! [[ ${IN_FILE} ]]; then
    echo "Input KML not provided."
    exit 1
fi

OUT_PATH=$2  # Output folder path.

if ! [[ ${OUT_PATH} ]]; then
    echo "Output path not provided."
    exit 1
fi

debug=false

if [[ "$3" == '-d' ]]; then
  debug=true
  echo 'Debug flag set to true.'
fi

docker build \
    -f monitoring/uss_qualifier/Dockerfile \
    -t interuss/uss_qualifier \
    --build-arg version="$(scripts/git/commit.sh)" \
    monitoring

docker run -i  -t --name flight_state_from_kml \
  --rm \
  --tty \
  -e PYTHONBUFFERED=1 \
  -v "${IN_FILE}:/app/kml-input/${IN_FILE}" \
  -v "${OUT_PATH}:/app/flight-states" \
  interuss/rid_qualifier \
  python rid/simulator/flight_state_from_kml.py -f "/app/kml-input/${IN_FILE}" -o /app/flight-states -d ${debug}
