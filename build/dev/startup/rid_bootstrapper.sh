#!/bin/sh

# This startup script is meant to be invoked from within a Docker container
# started by docker-compose_dss.yaml, not on a local system.

if ls /var/local-dss-data/rid_bootstrap > /dev/null 2>&1; then
  echo "RID DB is already set up; no need to run bootstrapper."
  exit 0
else
  echo "Allowing time for CRDB node to come up..."
  sleep 3

  echo "Bootstrapping RID DB..."
  /usr/bin/db-manager \
    --schemas_dir /db-schemas/defaultdb \
    --db_version "latest" \
    --cockroach_host local-dss-crdb

  echo "RID DB bootstrapping complete; notifying other containers..."
  touch /var/local-dss-data/rid_bootstrap
fi
