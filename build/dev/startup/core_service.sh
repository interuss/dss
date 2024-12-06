#!/bin/sh

# This startup script is meant to be invoked from within a Docker container
# started by docker-compose_dss.yaml, not on a local system.

DEBUG_ON=${1:-0}

# POSIX compliant test to check if with-yugabyte profile is enabled.
if [ "${COMPOSE_PROFILES#*"with-yugabyte"}" != "${COMPOSE_PROFILES}" ]; then
  echo "Using Yugabyte"
  DATASTORE_CONNECTION="-cockroach_host local-dss-ybdb -cockroach_user yugabyte --cockroach_port 5433"
else
  echo "Using CockroachDB"
  DATASTORE_CONNECTION="-cockroach_host local-dss-crdb"
fi

if [ "$DEBUG_ON" = "1" ]; then
  echo "Debug Mode: on"

  # Linter is disabled to properly unwrap $DATASTORE_CONNECTION.
  # shellcheck disable=SC2086
  dlv --headless --listen=:4000 --api-version=2 --accept-multiclient exec --continue /usr/bin/core-service -- ${DATASTORE_CONNECTION} \
  -public_key_files /var/test-certs/auth2.pem \
  -log_format console \
  -dump_requests \
  -addr :8082 \
  -accepted_jwt_audiences localhost,host.docker.internal,local-dss-core-service,dss_sandbox-local-dss-core-service-1,core-service \
  -enable_scd \
  -allow_http_base_urls
else
  echo "Debug Mode: off"

  # Linter is disabled to properly unwrap $DATASTORE_CONNECTION.
  # shellcheck disable=SC2086
  /usr/bin/core-service ${DATASTORE_CONNECTION} \
  -public_key_files /var/test-certs/auth2.pem \
  -log_format console \
  -dump_requests \
  -addr :8082 \
  -accepted_jwt_audiences localhost,host.docker.internal,local-dss-core-service,dss_sandbox-local-dss-core-service-1,core-service \
  -enable_scd \
  -allow_http_base_urls
fi
