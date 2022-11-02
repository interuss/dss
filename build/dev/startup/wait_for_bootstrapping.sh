#!/bin/sh

# This startup script is meant to be invoked from within a Docker container
# started by docker-compose_dss.yaml, not on a local system.

while [ ! -e /var/local-dss-data/rid_bootstrap ] || [ ! -e /var/local-dss-data/scd_bootstrap ]; do
  echo "Waiting on DB bootstrapping..."
  if [ ! -e /var/local-dss-data/rid_bootstrap ]; then
    echo "  RID pending."
  fi
  if [ ! -e /var/local-dss-data/scd_bootstrap ]; then
    echo "  SCD pending."
  fi
  sleep 3
done
echo "DB bootstrap complete; starting core service..."
