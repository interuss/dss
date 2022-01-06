#!/bin/sh

# This startup script is meant to be invoked from within a Docker container
# started by docker-compose_dss.yaml, not on a local system.

/startup/wait_for_bootstrapping.sh

DEBUG_ON=${1:-0}

if [ "$DEBUG_ON" = "1" ]; then
  echo "Debug Mode: on"

  dlv --headless --listen=:4000 --api-version=2 --accept-multiclient exec --continue /usr/bin/core-service -- \
  -cockroach_host local-dss-crdb \
  -public_key_files /var/test-certs/auth2.pem \
  -reflect_api \
  -log_format console \
  -dump_requests \
  -accepted_jwt_audiences localhost,host.docker.internal,local-gateway,dss_sandbox_local-dss-http-gateway_1 \
  -enable_scd \
  -enable_http
else
  echo "Debug Mode: off"

  /usr/bin/core-service \
  -cockroach_host local-dss-crdb \
  -public_key_files /var/test-certs/auth2.pem \
  -reflect_api \
  -log_format console \
  -dump_requests \
  -accepted_jwt_audiences localhost,host.docker.internal,local-gateway,dss_sandbox_local-dss-http-gateway_1 \
  -enable_scd \
  -enable_http
fi

