#!/usr/bin/env bash

set -eo pipefail

OS=$(uname)
if [[ $OS == "Darwin" ]]; then
	# OSX uses BSD readlink
	BASEDIR="$(dirname "$0")/.."
else
	BASEDIR=$(readlink -e "$(dirname "$0")/..")
fi

echo "Run Post migration Setup"
CRDB_MIGRATION_CONTAINER="dss-crdb-for-migration-testing"

if [ "$( docker container inspect -f '{{.State.Status}}' "$CRDB_MIGRATION_CONTAINER" )" == "running" ]; then
    echo "$CRDB_MIGRATION_CONTAINER available!"
else
    echo "Error: $CRDB_MIGRATION_CONTAINER not running. Execute 'clear_db.sh and migrate_db.sh' before running post_migration_e2e.sh";
    exit 1;
fi

echo "Bootstrapping SCD Database tables"
docker run --rm --name scd-db-manager \
	--link $CRDB_MIGRATION_CONTAINER:crdb \
	-v "$(pwd)/build/deploy/db_schemas/scd:/db-schemas/scd" \
	local-db-manager \
	--schemas_dir db-schemas/scd \
	--db_version "latest" \
	--cockroach_host crdb

sleep 1
echo " ------------ GRPC BACKEND ---------------- "
echo "Cleaning up any pre-existing grpc-backend container"
docker rm -f grpc-backend-for-testing &> /dev/null || echo "No grpc backend to clean up"

echo "Starting grpc backend on :8081"
docker run -d --name grpc-backend-for-testing \
	--link $CRDB_MIGRATION_CONTAINER:crdb \
	-v "$(pwd)/build/test-certs/auth2.pem:/app/test.crt" \
	local-interuss-dss-image \
	grpc-backend \
	--cockroach_host crdb \
	-public_key_files /app/test.crt \
	-reflect_api \
	-log_format console \
	-dump_requests \
	-accepted_jwt_audiences local-gateway \
	-enable_scd	\
	-enable_http

sleep 1
echo " ------------- HTTP GATEWAY -------------- "
echo "Cleaning up any pre-existing http-gateway container"
docker rm -f http-gateway-for-testing &> /dev/null || echo "No http gateway to clean up"

echo "Starting http-gateway on :8082"
docker run -d --name http-gateway-for-testing -p 8082:8082 \
	--link grpc-backend-for-testing:grpc \
	local-interuss-dss-image \
	http-gateway \
	-grpc-backend grpc:8081 \
	-addr :8082 \
	-trace-requests \
	-enable_scd

echo " -------------- DUMMY OAUTH -------------- "
echo "Building dummy-oauth server container"
docker build --rm -f cmds/dummy-oauth/Dockerfile . -t local-dummy-oauth > dummy-oauth-build.log

echo "Cleaning up any pre-existing dummy-oauth container"
docker rm -f dummy-oauth-for-testing &> /dev/null || echo "No dummy oauth to clean up"

echo "Starting mock oauth server on :8085"
docker run -d --name dummy-oauth-for-testing -p 8085:8085 \
	-v "$(pwd)/build/test-certs/auth2.key:/app/test.key" \
	local-dummy-oauth \
	-private_key_file /app/test.key

sleep 1
echo " -------------- PYTEST -------------- "
echo "Building Integration Test container"
docker build -q --rm -f monitoring/prober/Dockerfile monitoring -t e2e-test

echo "Finally Begin Testing"
docker run --link dummy-oauth-for-testing:oauth \
	--link http-gateway-for-testing:local-gateway \
	-v "${RESULTFILE}:/app/test_result" \
	e2e-test \
	"${1:-.}" \
	-rsx \
	--junitxml=/app/test_result \
	--dss-endpoint http://local-gateway:8082 \
	--rid-auth "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
	--scd-auth1 "DummyOAuth(http://oauth:8085/token,sub=fake_uss)" \
	--scd-auth2 "DummyOAuth(http://oauth:8085/token,sub=fake_uss2)"	\
	--scd-api-version 0.3.17

echo "Cleaning up http-gateway container"
docker stop http-gateway-for-testing > /dev/null
test "$(docker inspect http-gateway-for-testing --format='{{.State.ExitCode}}')" = 0

echo "Cleaning up grpc-backend container"
docker stop grpc-backend-for-testing > /dev/null
test "$(docker inspect grpc-backend-for-testing --format='{{.State.ExitCode}}')" = 0
