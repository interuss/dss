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

echo '################################################################################'
echo '## NOTE: A prerequisite for running this command locally is to have a running ##'
echo '## instance of the uss_qualifier RID system mock (rid/mock/run_locally.sh)    ##'
echo '################################################################################'

CONFIG_LOCATION="monitoring/uss_qualifier/config_run_locally.json"
CONFIG='--config config_run_locally.json'

AUTH='--auth NoAuth()'

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
    ]
  }
}' > ${CONFIG_LOCATION}

RID_QUALIFIER_OPTIONS="$AUTH $CONFIG"

REPORT_FILE="$(pwd)/monitoring/uss_qualifier/report.json"
# report.json must already exist to share correctly with the Docker container
touch "${REPORT_FILE}"
#
#docker build \
#    -f monitoring/uss_qualifier/Dockerfile \
#    -t interuss/uss_qualifier \
#    --build-arg version="$(scripts/git/commit.sh)" \
#    monitoring

if [ "$CI" == "true" ]; then
  docker_args="--add-host host.docker.internal:host-gateway" # Required to reach other containers in Ubuntu (used for Github Actions)
else
  docker_args="-it"
fi


# shellcheck disable=SC2086
docker run ${docker_args} --name uss_qualifier \
  --rm \
  -e RID_QUALIFIER_OPTIONS="${RID_QUALIFIER_OPTIONS}" \
  -e PYTHONBUFFERED=1 \
  -v "${REPORT_FILE}:/app/monitoring/uss_qualifier/report.json" \
  -v "$(pwd):/app" \
  interuss/uss_qualifier \
  python main.py $RID_QUALIFIER_OPTIONS

rm ${CONFIG_LOCATION}
