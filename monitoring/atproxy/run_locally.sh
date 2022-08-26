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

monitoring/build.sh || exit 1

CLIENT_BASIC_AUTH="local_client:local_client"
PUBLIC_KEY="/var/test-certs/auth2.pem"
AUD=${MOCK_USS_TOKEN_AUDIENCE:-localhost,host.docker.internal}

PORT=8075

if [ "$CI" == "true" ]; then
  docker_args="--add-host host.docker.internal:host-gateway" # Required to reach other containers in Ubuntu (used for Github Actions)
else
  docker_args=""
fi

# shellcheck disable=SC2086
docker run ${docker_args} --name atproxy \
  --rm \
  -e ATPROXY_CLIENT_BASIC_AUTH="${CLIENT_BASIC_AUTH}" \
  -e ATPROXY_PUBLIC_KEY="${PUBLIC_KEY}" \
  -e ATPROXY_TOKEN_AUDIENCE="${AUD}" \
  -p ${PORT}:5000 \
  -v "$(pwd)/build/test-certs:/var/test-certs:ro" \
  "$@" \
  interuss/monitoring \
  atproxy/start.sh
