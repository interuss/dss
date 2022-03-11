#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
"${SCRIPT_DIR}/run_locally_base.sh"

AUTH="DummyOAuth(http://host.docker.internal:8085/token,uss1)"
DSS="http://host.docker.internal:8082"
PUBLIC_KEY="/var/test-certs/auth2.pem"
AUD=${MOCK_USS_TOKEN_AUDIENCE:-localhost,host.docker.internal}

PORT=8074
BASE_URL="http://${MOCK_USS_TOKEN_AUDIENCE:-host.docker.internal}:${PORT}"

docker run --name mock_uss_scdsc \
  --rm \
  -e MOCK_USS_AUTH_SPEC="${AUTH}" \
  -e MOCK_USS_DSS_URL="${DSS}" \
  -e MOCK_USS_PUBLIC_KEY="${PUBLIC_KEY}" \
  -e MOCK_USS_TOKEN_AUDIENCE="${AUD}" \
  -e MOCK_USS_BASE_URL="${BASE_URL}" \
  -e MOCK_USS_SERVICES="scdsc" \
  -p ${PORT}:5000 \
  -v "${SCRIPT_DIR}/../../build/test-certs:/var/test-certs:ro" \
  $* \
  local-interuss/mock_uss \
  gunicorn \
    --preload \
    --workers=2 \
    --bind=0.0.0.0:5000 \
    monitoring.mock_uss:webapp
