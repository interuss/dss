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
AUD="host.docker.internal,localhost"
PORT=8073

docker build \
  -t local-interuss/mock_riddp \
  -f monitoring/mock_riddp/Dockerfile \
  --build-arg version="$(scripts/git/commit.sh)" \
  monitoring \
  || exit 1

docker run --name mock_riddp \
  --rm \
  -e MOCK_RIDDP_AUTH_SPEC="${AUTH}" \
  -e MOCK_RIDDP_DSS_URL="${DSS}" \
  -e MOCK_RIDDP_PUBLIC_KEY="${PUBLIC_KEY}" \
  -e MOCK_RIDDP_TOKEN_AUDIENCE="${AUD}" \
  -p ${PORT}:5000 \
  -v "$(pwd)/build/test-certs:/var/test-certs:ro" \
  local-interuss/mock_riddp \
  gunicorn \
    --preload \
    --workers=1 \
    --bind=0.0.0.0:5000 \
    monitoring.mock_riddp:webapp
