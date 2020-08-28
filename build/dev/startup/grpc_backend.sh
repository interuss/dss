#!/bin/sh

# This startup script is meant to be invoked from within a Docker container
# started by docker-compose_dss.yaml, not on a local system.

/startup/wait_for_bootstrapping.sh

/usr/bin/grpc-backend \
  -cockroach_host local-dss-crdb \
  -public_key_files /var/test-certs/auth2.pem \
  -reflect_api \
  -log_format console \
  -dump_requests \
  -accepted_jwt_audiences localhost \
  -enable_scd
