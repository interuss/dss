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

CONFIG_LOCATION="monitoring/rid_qualifier/config.json"

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

CONFIG='--config config.json'

RID_QUALIFIER_OPTIONS="$AUTH $CONFIG"

# report.json must already exist to share correctly with the Docker container
touch $(pwd)/monitoring/rid_qualifier/report.json

docker build \
    -f monitoring/rid_qualifier/Dockerfile \
    -t interuss/dss/rid_qualifier \
    --build-arg version=`scripts/git/commit.sh` \
    monitoring

docker run --name rid_qualifier \
  --rm \
  --tty \
  -e RID_QUALIFIER_OPTIONS="${RID_QUALIFIER_OPTIONS}" \
  -e PYTHONBUFFERED=1 \
  -v $(pwd)/monitoring/rid_qualifier/report.json:/app/monitoring/rid_qualifier/report.json \
  -v $(pwd)/${CONFIG_LOCATION}:/app/${CONFIG_LOCATION} \
  interuss/dss/rid_qualifier \
  python rid_qualifier_entry.py $RID_QUALIFIER_OPTIONS

rm ${CONFIG_LOCATION}
