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

echo '################################################################################'
echo '## NOTE: Prerequisites to run this command are:                               ##'
echo '## 1. Local DSS instance + Dummy OAuth server (/build/dev/run_locally.sh)     ##'
echo '## 2. Local mock RID service provider (/monitoring/mock_ridsp/run_locally.sh) ##'
echo '## 3. Local mock RID display provider (/monitoring/mock_riddp/run_locally.sh) ##'
echo '################################################################################'

CONFIG_LOCATION="monitoring/uss_qualifier/config_test_fully_mocked_local_system.json"
CONFIG='--config config_test_fully_mocked_local_system.json'

AUTH='--auth DummyOAuth(http://host.docker.internal:8085/token,sub=testing_uss)'

echo '{
  "locale": "che",
  "injection_targets": [
    {
      "name": "uss1",
      "injection_base_url": "http://host.docker.internal:8071/injection"
    }
  ],
  "observers": [
    {
      "name": "uss2",
      "observation_base_url": "http://host.docker.internal:8073/observation"
    }
  ]
}' > ${CONFIG_LOCATION}

FLIGHT_RECORDS_PATH='--flight-records /app/monitoring/uss_qualifier/test_flights'

RID_QUALIFIER_OPTIONS="$AUTH $CONFIG $FLIGHT_RECORDS_PATH"

# report.json must already exist to share correctly with the Docker container
touch "$(pwd)/monitoring/uss_qualifier/report.json"

docker build \
    -f monitoring/uss_qualifier/Dockerfile \
    -t interuss/uss_qualifier \
    --build-arg version="$(scripts/git/commit.sh)" \
    monitoring

# shellcheck disable=SC2086
docker run --name uss_qualifier \
  --rm \
  --tty \
  -e RID_QUALIFIER_OPTIONS="${RID_QUALIFIER_OPTIONS}" \
  -e PYTHONBUFFERED=1 \
  -v "$(pwd)/monitoring/uss_qualifier/report.json:/app/monitoring/uss_qualifier/report.json" \
  -v "$(pwd)/monitoring/uss_qualifier/test_flights:/app/monitoring/uss_qualifier/test_flights" \
  -v "$(pwd)/${CONFIG_LOCATION}:/app/${CONFIG_LOCATION}" \
  interuss/uss_qualifier \
  python main.py $RID_QUALIFIER_OPTIONS

rm ${CONFIG_LOCATION}
