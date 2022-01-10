#!/usr/bin/env bash

set -eo pipefail

echo " ---------------- START DATABASE -------------------- "
docker rm -f dss-crdb-for-migration-testing &> /dev/null || echo "No CRDB to clean up"
echo "Starting CRDB container"
docker run -d --rm --name dss-crdb-for-migration-testing \
	-p 26257:26257 \
	-p 8080:8080 \
  cockroachdb/cockroach:v21.2.3 start-single-node \
  --insecure > /dev/null
