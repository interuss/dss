#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
"${SCRIPT_DIR}/run_locally_base.sh"

AUTH="DummyOAuth(http://host.docker.internal:8085/token,uss2)"
DSS="http://host.docker.internal:8082"
PUBLIC_KEY="/var/test-certs/auth2.pem"
AUD=${MOCK_USS_TOKEN_AUDIENCE:-localhost,host.docker.internal}

PORT=8077
BASE_URL="http://${MOCK_USS_TOKEN_AUDIENCE:-host.docker.internal}:${PORT}"

ATPROXY_PORT=8075
ATPROXY_BASIC_AUTH="local_client:local_client"
ATPROXY_BASE_URL="http://${MOCK_USS_TOKEN_AUDIENCE:-host.docker.internal}:${ATPROXY_PORT}"

if [ "$CI" == "true" ]; then
  docker_args="--add-host host.docker.internal:host-gateway" # Required to reach other containers in Ubuntu (used for Github Actions)
else
  docker_args=""
fi

# shellcheck disable=SC2086
docker run ${docker_args} --name mock_uss_atproxy_client \
  --rm \
  -e MOCK_USS_AUTH_SPEC="${AUTH}" \
  -e MOCK_USS_DSS_URL="${DSS}" \
  -e MOCK_USS_PUBLIC_KEY="${PUBLIC_KEY}" \
  -e MOCK_USS_TOKEN_AUDIENCE="${AUD}" \
  -e MOCK_USS_BASE_URL="${BASE_URL}" \
  -e MOCK_USS_ATPROXY_BASE_URL="${ATPROXY_BASE_URL}" \
  -e MOCK_USS_ATPROXY_BASIC_AUTH="${ATPROXY_BASIC_AUTH}" \
  -e MOCK_USS_SERVICES="atproxy_client,scdsc,ridsp,riddp" \
  -p ${PORT}:5000 \
  -v "${SCRIPT_DIR}/../../build/test-certs:/var/test-certs:ro" \
  "$@" \
  interuss/monitoring \
  mock_uss/start.sh
