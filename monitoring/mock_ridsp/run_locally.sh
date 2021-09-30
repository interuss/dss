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
AUD=${MOCK_RIDSP_TOKEN_AUDIENCE:-localhost}
BASE_URL="http://${AUD}:8071/ridsp"
PORT=8071

docker build \
  -t local-interuss/mock_ridsp \
  -f monitoring/mock_ridsp/Dockerfile \
  --build-arg version="$(scripts/git/commit.sh)" \
  monitoring \
  || exit 1

docker run --name mock_ridsp \
  --rm \
  -e MOCK_RIDSP_AUTH_SPEC="${AUTH}" \
  -e MOCK_RIDSP_DSS_URL="${DSS}" \
  -e MOCK_RIDSP_PUBLIC_KEY="${PUBLIC_KEY}" \
  -e MOCK_RIDSP_TOKEN_AUDIENCE="${AUD}" \
  -e MOCK_RIDSP_BASE_URL="${BASE_URL}" \
  -p ${PORT}:5000 \
  -v "$(pwd)/build/test-certs:/var/test-certs:ro" \
  local-interuss/mock_ridsp \
  gunicorn \
    --preload \
    --workers=1 \
    --bind=0.0.0.0:5000 \
    monitoring.mock_ridsp:webapp
