#!/bin/sh

# This startup script is meant to be invoked from within a Docker container
# started by docker-compose_dss.yaml, not on a local system.

/startup/wait_for_bootstrapping.sh

echo "Allowing time for Core Service to come up..."
sleep 3

echo "Starting HTTP gateway..."
/usr/bin/http-gateway \
  -core-service local-dss-core-service:8081 \
  -addr :8082 \
  -trace-requests \
  -enable_scd
