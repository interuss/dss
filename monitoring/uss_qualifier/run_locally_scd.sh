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

CONFIG_LOCATION="monitoring/uss_qualifier/config_run_locally_scd.json"
CONFIG='--config config_run_locally_scd.json'

AUTH='--auth DummyOAuth(http://host.docker.internal:8085/token,uss_qualifier)'

echo '{
    "locale": "CHE",
    "scd": {
        "injection_targets": [
            {
                "name": "uss1",
                "injection_base_url": "http://host.docker.internal:8074/scdsc"
            },
            {
                "name": "uss2",
                "injection_base_url": "http://host.docker.internal:8074/scdsc"
            }
        ],
        "dss_base_url": "http://host.docker.internal:8082"
    }
}' > ${CONFIG_LOCATION}

QUALIFIER_OPTIONS="$AUTH $CONFIG"

REPORT_FILE="$(pwd)/monitoring/uss_qualifier/report_scd.json"
# Report file must already exist to share correctly with the Docker container
touch "${REPORT_FILE}"

docker build \
    -f monitoring/uss_qualifier/Dockerfile \
    -t interuss/uss_qualifier \
    --build-arg version="$(scripts/git/commit.sh)" \
    --build-arg qualifier_version="$(scripts/git/version.sh uss_qualifier --long)" \
    monitoring

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
  -v "${REPORT_FILE}:/app/monitoring/uss_qualifier/report_scd.json" \
  -v "$(pwd):/app" \
  interuss/uss_qualifier \
  python main.py $QUALIFIER_OPTIONS

rm ${CONFIG_LOCATION}
