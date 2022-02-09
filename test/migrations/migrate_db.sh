#!/usr/bin/env bash

set -eo pipefail

db_name=$1
version=$2
crdb_name=${3:-"dss-crdb-for-migration-testing"}

if [[ -n "$4" ]]
then
  network_flag="--network $4"
else
  network_flag=""
fi

echo "network flag: ${network_flag}"

echo "Migrating ${db_name} database to version ${version}"
echo "crdb server: ${crdb_name} dss network ${network_flag}"

echo " -------------- BOOTSTRAP ----------------- "
echo "Building db-manager container for testing"
docker build --rm -f cmds/db-manager/Dockerfile . -t local-db-manager > db-manager-build.log

echo " ---------------- MIGRATE DATABASE -------------------- "
echo "Migrating ${db_name} database to version ${version}"
docker run --rm --name migration-testing-db-manager \
  --link "${crdb_name}":crdb \
  "$network_flag" \
  -v "$(pwd)/build/deploy/db_schemas/${db_name}:/db-schemas/${db_name}" \
  local-db-manager \
  --schemas_dir db-schemas/"${db_name}" \
  --db_version "${version}" \
  --cockroach_host crdb
