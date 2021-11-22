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

CONFIG_LOCATION="monitoring/rid_qualifier/config.json"

AUTH='--auth NoAuth()'
# NB: A prerequisite to run this command locally is to have a running instance of the rid_qualifier/mock (/monitoring/rid_qualifier/mock/run_locally.sh)

echo '{
  "locale": "che",
  "injection_targets": [
    {
      "name": "uss1",
      "injection_base_url": "http://host.docker.internal:8070/sp/uss1"
    },
    {
      "name": "uss2",
      "injection_base_url": "http://host.docker.internal:8070/sp/uss2"
    }
  ],
  "observers": [
    {
      "name": "uss2",
      "observation_base_url": "http://host.docker.internal:8070/dp/uss2"
    },
    {
      "name": "uss3",
      "observation_base_url": "http://host.docker.internal:8070/dp/uss3"
    }
  ]
}' > ${CONFIG_LOCATION}

CONFIG='--config config.json'

RID_QUALIFIER_OPTIONS="$AUTH $CONFIG"

# report.json must already exist to share correctly with the Docker container
touch "$(pwd)/monitoring/rid_qualifier/report.json"

docker build \
    -f monitoring/rid_qualifier/Dockerfile \
    -t interuss/dss/rid_qualifier \
    --build-arg version="$(scripts/git/commit.sh)" \
    monitoring

if [ "$CI" == "true" ]; then
  docker_args="--add-host host.docker.internal:host-gateway" # Required to reach other containers in Ubuntu (used for Github Actions)
else
  docker_args="-it"
fi

# shellcheck disable=SC2086
docker run ${docker_args} --name rid_qualifier \
  --rm \
  -e RID_QUALIFIER_OPTIONS="${RID_QUALIFIER_OPTIONS}" \
  -e PYTHONBUFFERED=1 \
  -v "$(pwd)/monitoring/rid_qualifier/report.json:/app/monitoring/rid_qualifier/report.json" \
  -v "$(pwd)/${CONFIG_LOCATION}:/app/${CONFIG_LOCATION}" \
  interuss/dss/rid_qualifier \
  python rid_qualifier_entry.py $RID_QUALIFIER_OPTIONS

rm ${CONFIG_LOCATION}
