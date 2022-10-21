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
cd "${BASEDIR}/../.." || exit 1

echo '#########################################################################'
echo '## NOTE: A prerequisite for running this command locally is to have    ##'
echo '## a running instance of the mock_uss with SCD enabled                 ##'
echo '## (../mock_uss/run_locally_scdsc.sh) including related dependencies.  ##'
echo '#########################################################################'

monitoring/build.sh || exit 1

CONFIG_LOCATION="monitoring/uss_qualifier/config_run_locally_scd.json"
CONFIG='--config config_run_locally_scd.json'

AUTH_SPEC='DummyOAuth(http://host.docker.internal:8085/token,uss_qualifier)'

echo '{
    "resources": {
        "resource_declarations": {
          "utm_auth": {
            "resource_type": "communications.AuthAdapterResource",
            "specification": {
              "environment_variable_containing_auth_spec": "AUTH_SPEC"
            }
          },
          "flight_planners": {
            "resource_type": "flight_planning.FlightPlannersResource",
            "dependencies": {
              "auth_adapter": "utm_auth"
            },
            "specification": {
              "flight_planners": [
                {
                    "participant_id": "uss1",
                    "injection_base_url": "http://host.docker.internal:8074/scdsc"
                },
                {
                    "participant_id": "uss2",
                    "injection_base_url": "http://host.docker.internal:8074/scdsc"
                }
              ]
            }
          },
          "dss_instance": {
            "resource_type": "astm.f3548.v21.DSSInstanceResource",
            "dependencies": {
              "auth_adapter": "utm_auth"
            },
            "specification": {
              "participant_id": "uss1",
              "base_url": "http://host.docker.internal:8082"
            }
          },
          "invalid_flight_auth_flights": {
            "resource_type": "flight_planning.FlightIntentsResource",
            "specification": {
              "planning_time": "0:05:00",
              "json_file_source": {
                "path": "./test_data/che/flight_intents/invalid_flight_auths.json"
              }
            }
          },
          "conflicting_flights": {
            "resource_type": "flight_planning.FlightIntentsResource",
            "specification": {
              "planning_time": "0:05:00",
              "json_file_source": {
                "path": "./test_data/che/flight_intents/conflicting_flights.json"
              }
            }
          },
          "priority_preemption": {
            "resource_type": "flight_planning.FlightIntentsResource",
            "specification": {
              "planning_time": "0:05:00",
              "json_file_source": {
                "path": "./test_data/che/flight_intents/priority_preemption.json"
              }
            }
          }
        }
    }
}' > ${CONFIG_LOCATION}

QUALIFIER_OPTIONS="$CONFIG"

REPORT_FILE="$(pwd)/monitoring/uss_qualifier/report_scd.json"
# Report file must already exist to share correctly with the Docker container
touch "${REPORT_FILE}"

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
  -e AUTH_SPEC=${AUTH_SPEC} \
  -v "${REPORT_FILE}:/app/monitoring/uss_qualifier/report_scd.json" \
  -v "$(pwd):/app" \
  -w /app/monitoring/uss_qualifier \
  interuss/monitoring \
  python main.py $QUALIFIER_OPTIONS

rm ${CONFIG_LOCATION}
