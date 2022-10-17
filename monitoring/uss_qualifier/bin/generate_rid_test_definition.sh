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

CONFIG_LOCATION="monitoring/uss_qualifier/config_run_locally.json"
CONFIG='--config config_run_locally.json'

echo '{
    "locale": "CHE",
    "resources": {
      "resource_declarations": {
        "adjacent_circular_flights_data": {
          "resource_type": "netrid.FlightDataResource",
          "specification": {
            "adjacent_circular_flights_simulation_source": {}
          }
        },
        "adjacent_circular_storage_config": {
          "resource_type": "netrid.FlightDataStorageResource",
          "specification": {
            "flight_record_collection_path": "./test_data/che/netrid/circular_flights.json"
          }
        },
        "kml_flights_data": {
          "resource_type": "netrid.FlightDataResource",
          "specification": {
            "kml_file_source": {
              "kml_path": "./test_data/usa/netrid/dcdemo.kml"
            }
          }
        },
        "kml_storage_config": {
          "resource_type": "netrid.FlightDataStorageResource",
          "specification": {
            "flight_record_collection_path": "./test_data/usa/netrid/dcdemo_flights.json"
          }
        }
      }
    },
    "suite": {
      "suite_type": "interuss.generate_test_data",
      "resources": {
        "adjacent_circular_flights_data": "adjacent_circular_flights_data",
        "adjacent_circular_storage_config": "adjacent_circular_storage_config",
        "kml_flights_data": "kml_flights_data",
        "kml_storage_config": "kml_storage_config"
      }
    }
}' > ${CONFIG_LOCATION}

QUALIFIER_OPTIONS="--auth NA $CONFIG"

if [ "$CI" == "true" ]; then
  docker_args="--add-host host.docker.internal:host-gateway" # Required to reach other containers in Ubuntu (used for Github Actions)
else
  docker_args="-it"
fi

# shellcheck disable=SC2086
docker run ${docker_args} --name uss_qualifier \
  --rm \
  -e QUALIFIER_OPTIONS="${QUALIFIER_OPTIONS}" \
  -e PYTHONBUFFERED=1 \
  -v "$(pwd):/app" \
  -w /app/monitoring/uss_qualifier \
  interuss/monitoring \
  python main.py $QUALIFIER_OPTIONS

rm ${CONFIG_LOCATION}
