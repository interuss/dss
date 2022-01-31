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

# This sample assumes that a local DSS instance is available similar to the one
# produced with /build/dev/run_locally.sh
AUTH="DummyOAuth(http://host.docker.internal:8085/token,uss1)"
DSS="http://host.docker.internal:8082"
PUBLIC_KEY="/var/test-certs/auth2.pem"
AUD=${MOCK_USS_TOKEN_AUDIENCE:-localhost,host.docker.internal}
BASE_URL="http://${MOCK_USS_TOKEN_AUDIENCE:-host.docker.internal}:8071"
PORT=8071

docker build \
  -t local-interuss/mock_uss \
  -f monitoring/mock_uss/Dockerfile \
  --build-arg version="$(scripts/git/commit.sh)" \
  monitoring \
  || exit 1

docker run --name mock_uss \
  --rm \
  -e MOCK_USS_AUTH_SPEC="${AUTH}" \
  -e MOCK_USS_DSS_URL="${DSS}" \
  -e MOCK_USS_PUBLIC_KEY="${PUBLIC_KEY}" \
  -e MOCK_USS_TOKEN_AUDIENCE="${AUD}" \
  -e MOCK_USS_BASE_URL="${BASE_URL}" \
  -e MOCK_USS_SERVICES="ridsp" \
  -p ${PORT}:5000 \
  -v "$(pwd)/build/test-certs:/var/test-certs:ro" \
  local-interuss/mock_uss \
  gunicorn \
    --preload \
    --workers=1 \
    --bind=0.0.0.0:5000 \
    monitoring.mock_uss:webapp
